package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"github.com/adamluzsi/frameless"
	"github.com/adamluzsi/frameless/iterators"
	"github.com/pkg/errors"

	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/toggler/services/rollouts"
	"github.com/adamluzsi/toggler/services/security"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

func New(connstr string) (*Redis, error) {
	r := &Redis{}
	redisClientOpt, err := redis.ParseURL(connstr)
	if err != nil {
		return nil, err
	}
	r.client = redis.NewClient(redisClientOpt)
	r.keyMapping = map[string]string{
		reflects.FullyQualifiedName(security.Token{}):       "tokens",
		reflects.FullyQualifiedName(rollouts.Pilot{}):       "pilots",
		reflects.FullyQualifiedName(rollouts.FeatureFlag{}): "feature_flags",
		reflects.FullyQualifiedName(resources.TestEntity{}): "test_entities",
	}
	return r, nil
}

type Redis struct {
	client     *redis.Client
	keyMapping map[string]string
}

func (r *Redis) FindFlagsByName(ctx context.Context, names ...string) frameless.Iterator {
	flags := r.FindAll(ctx, rollouts.FeatureFlag{})

	nameIndex := make(map[string]struct{})

	for _, name := range names {
		nameIndex[name] = struct{}{}
	}

	flagsByName := iterators.Filter(flags, func(flag frameless.Entity) bool {
		_, ok := nameIndex[flag.(rollouts.FeatureFlag).Name]
		return ok
	})

	return flagsByName
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) FindByID(ctx context.Context, ptr interface{}, ID string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	reply := r.client.WithContext(ctx).HGet(r.hKey(ptr), ID)

	err := reply.Err()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	bs, err := reply.Bytes()

	if err != nil {
		return false, err
	}

	return true, r.unmarshal(ptr, bs)
}

func (r *Redis) DeleteByID(ctx context.Context, Type interface{}, ID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	reply := r.client.WithContext(ctx).HDel(r.hKey(Type), ID)
	if err := reply.Err(); err != nil && err != redis.Nil {
		return err
	}

	success, err := reply.Result()
	if err != nil {
		return err
	}

	if success != 1 {
		return frameless.ErrNotFound
	}

	return nil
}

func (r *Redis) Truncate(ctx context.Context, Type interface{}) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	hkey := r.hKey(Type)
	keys, err := r.client.WithContext(ctx).HKeys(hkey).Result()
	if err != nil {
		return err
	}

	for _, ID := range keys {
		if err := r.DeleteByID(ctx, Type, ID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Redis) Update(ctx context.Context, ptr interface{}) error {
	id, found := resources.LookupID(ptr)
	if !found {
		return frameless.ErrIDRequired
	}

	found, err := r.FindByID(ctx, reflects.New(ptr), id)
	if err != nil && err != frameless.ErrNotFound {
		return err
	}
	if !found {
		return frameless.ErrNotFound
	}

	bs, err := r.marshal(ptr)
	if err != nil {
		return err
	}

	return r.client.WithContext(ctx).HSet(r.hKey(ptr), id, bs).Err()
}

func (r *Redis) FindAll(ctx context.Context, Type interface{}) frameless.Iterator {
	if err := ctx.Err(); err != nil {
		return iterators.NewError(err)
	}

	valuesWithIDs, err := r.client.WithContext(ctx).HGetAll(r.hKey(Type)).Result()

	if err != nil {
		return iterators.NewError(err)
	}

	var values []interface{}

	for _, serialized := range valuesWithIDs {
		v := reflects.New(Type)
		if err := r.unmarshal(v, []byte(serialized)); err != nil {
			return iterators.NewError(err)
		}
		values = append(values, v)
	}

	return iterators.NewSlice(values)
}

func (r *Redis) FindFlagByName(ctx context.Context, name string) (*rollouts.FeatureFlag, error) {
	flags := r.FindAll(ctx, rollouts.FeatureFlag{})
	flagsByName := iterators.Filter(flags, func(flag frameless.Entity) bool {
		return flag.(rollouts.FeatureFlag).Name == name
	})
	var flag rollouts.FeatureFlag
	err := iterators.First(flagsByName, &flag)

	if err == frameless.ErrNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &flag, nil
}

func (r *Redis) FindFlagPilotByExternalPilotID(ctx context.Context, FeatureFlagID, ExternalPilotID string) (*rollouts.Pilot, error) {
	pilots := r.FindAll(ctx, rollouts.Pilot{})
	pilotsByIDs := iterators.Filter(pilots, func(pilot frameless.Entity) bool {
		p := pilot.(rollouts.Pilot)

		return p.FeatureFlagID == FeatureFlagID && p.ExternalID == ExternalPilotID
	})
	var p rollouts.Pilot
	err := iterators.First(pilotsByIDs, &p)

	if err == frameless.ErrNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (r *Redis) FindPilotsByFeatureFlag(ctx context.Context, ff *rollouts.FeatureFlag) frameless.Iterator {
	pilots := r.FindAll(ctx, rollouts.Pilot{})
	return iterators.Filter(pilots, func(pilot frameless.Entity) bool {
		return pilot.(rollouts.Pilot).FeatureFlagID == ff.ID
	})
}

func (r *Redis) FindPilotEntriesByExtID(ctx context.Context, pilotExtID string) rollouts.PilotEntries {
	pilots := r.FindAll(ctx, rollouts.Pilot{})
	return iterators.Filter(pilots, func(pilot frameless.Entity) bool {
		return pilot.(rollouts.Pilot).ExternalID == pilotExtID
	})
}

func (r *Redis) FindTokenBySHA512Hex(ctx context.Context, sha512hex string) (*security.Token, error) {
	tokens := r.FindAll(ctx, security.Token{})
	tokensBySHA512 := iterators.Filter(tokens, func(token frameless.Entity) bool {
		return token.(security.Token).SHA512 == sha512hex
	})
	var t security.Token
	err := iterators.First(tokensBySHA512, &t)

	if err == frameless.ErrNotFound {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &t, nil
}

func (r *Redis) Save(ctx context.Context, ptr interface{}) error {
	currentID, found := resources.LookupID(ptr)

	if !found {
		return frameless.ErrIDRequired
	}

	if currentID != "" {
		return errors.New(`id already set`)
	}

	switch e := ptr.(type) {
	case *rollouts.FeatureFlag:
		ff, err := r.FindFlagByName(ctx, e.Name)
		if err != nil {
			return err
		}
		if ff != nil {
			return errors.New(`flag uniq constrain violation for "name" attr`)
		}
	}

	id := uuid.New().String()

	if id == `` {
		return errors.New(`error happened during creating new id`)
	}

	if err := resources.SetID(ptr, id); err != nil {
		return err
	}

	bs, err := r.marshal(ptr)
	if err != nil {
		return err
	}

	// apparently the redis client is not really
	// worried if it received a canceled context.
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := r.client.WithContext(ctx).HSet(r.hKey(ptr), id, bs).Err(); err != nil {
		return err
	}

	return nil
}

//--------------------------------------------------------------------------------------------------------------------//

func (r *Redis) marshal(ptr interface{}) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(ptr)
	return buf.Bytes(), err
}

func (r *Redis) unmarshal(ptr interface{}, bs []byte) error {
	decoder := gob.NewDecoder(bytes.NewReader(bs))
	return decoder.Decode(ptr)
}

func (r *Redis) hKey(T interface{}) string {
	mappingKey := reflects.FullyQualifiedName(T)
	key, ok := r.keyMapping[mappingKey]
	if !ok {
		panic(fmt.Sprintf(`missing key mapping for redis storage implementation: %s`, mappingKey))
	}
	return key
}
