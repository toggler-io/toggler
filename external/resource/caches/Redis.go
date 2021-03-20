package caches

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/resources"
	"github.com/go-redis/redis"
	"github.com/toggler-io/toggler/external/resource/storages"
)

func NewRedisCacheStorage(connstr string) (*RedisCacheStorage, error) {
	redisClientOpt, err := redis.ParseURL(connstr)
	if err != nil {
		return nil, err
	}
	return &RedisCacheStorage{
		client:                    redis.NewClient(redisClientOpt),
		EntityTypeNameMappingFunc: defaultEntityTypeNameMapping,
		EnsureID:                  storages.EnsureID,
	}, nil
}

type RedisCacheStorage struct {
	client *redis.Client
	EntityTypeNameMappingFunc
	EnsureID func(ptr interface{}) error
}

func (r *RedisCacheStorage) Close() error {
	return r.client.Close()
}

func (r *RedisCacheStorage) Create(ctx context.Context, ptr interface{}) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := r.EnsureID(ptr); err != nil {
		return err
	}

	id, ok := resources.LookupID(ptr)
	if !ok {
		return fmt.Errorf(`creating without id is not supported`)
	}
	key := r.getKey(getT(ptr), id)

	found, err := r.FindByID(ctx, reflect.New(reflect.TypeOf(ptr).Elem()).Interface(), id)
	if err != nil {
		return err
	}
	if found {
		return fmt.Errorf(`%T/%v is already stored in the cache storage`, ptr, id)
	}

	bs, err := r.marshal(ptr)
	if err != nil {
		return err
	}

	return r.client.WithContext(ctx).Set(key, bs, 0).Err()
}

func (r *RedisCacheStorage) FindByID(ctx context.Context, ptr, id interface{}) (_found bool, _err error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}

	get := r.client.WithContext(ctx).Get(r.getKey(getT(ptr), id))

	if err := get.Err(); err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	bs, err := get.Bytes()
	if err != nil {
		return false, err
	}

	if err := r.unmarshal(bs, ptr); err != nil {
		return false, err
	}

	return true, nil
}

func (r *RedisCacheStorage) FindAll(ctx context.Context, T resources.T) iterators.Interface {
	if err := ctx.Err(); err != nil {
		return iterators.NewError(err)
	}

	keysRes := r.client.WithContext(ctx).Keys(fmt.Sprintf(`%s*`, r.getKeyPrefix(T)))

	if err := keysRes.Err(); err != nil {
		if err == redis.Nil {
			return iterators.NewEmpty()
		}
		return iterators.NewError(err)
	}

	keys, err := keysRes.Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return iterators.NewError(err)
	}

	pipe, sender := iterators.NewPipe()
	rt := reflect.TypeOf(T)

	go func() {
		defer sender.Close()
		for _, key := range keys {

			get := r.client.WithContext(ctx).Get(key)
			if err := get.Err(); err != nil {
				sender.Error(err)
				return
			}

			bs, err := get.Bytes()
			if err != nil {
				sender.Error(err)
				return
			}

			ptr := reflect.New(rt)
			if err := r.unmarshal(bs, ptr.Interface()); err != nil {
				sender.Error(err)
				return
			}

			if err := sender.Encode(ptr.Elem().Interface()); err != nil {
				return
			}
		}
	}()

	return pipe
}

func (r *RedisCacheStorage) Update(ctx context.Context, ptr interface{}) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	id := r.getID(ptr)
	found, err := r.FindByID(ctx, reflect.New(reflect.TypeOf(ptr).Elem()).Interface(), id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf(`not found: %T / %v`, ptr, id)
	}

	key := r.getKey(getT(ptr), id)
	bs, err := r.marshal(ptr)
	if err != nil {
		return err
	}
	return r.client.Set(key, bs, 0).Err()
}

func (r *RedisCacheStorage) DeleteByID(ctx context.Context, T, id interface{}) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	found, err := r.FindByID(ctx, reflect.New(reflect.TypeOf(T)).Interface(), id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf(`not found: %T / %v`, T, id)
	}

	return r.client.Del(r.getKey(T, id)).Err()
}

func (r *RedisCacheStorage) DeleteAll(ctx context.Context, T resources.T) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	keysRes := r.client.Keys(fmt.Sprintf(`%s*`, r.getKeyPrefix(T)))

	if err := keysRes.Err(); err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	keys, err := keysRes.Result()
	if err == redis.Nil {
		return nil
	}
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}

	return r.client.Del(keys...).Err()
}

func (r *RedisCacheStorage) UpsertMany(ctx context.Context, ptrs ...interface{}) error {
	return r.client.WithContext(ctx).Watch(func(tx *redis.Tx) error {
		for _, ptr := range ptrs {
			if err := r.EnsureID(ptr); err != nil {
				return err
			}

			bs, err := r.marshal(ptr)
			if err != nil {
				return err
			}

			tx.Set(r.getKey(getT(ptr), r.getID(ptr)), bs, 0)
		}

		return nil
	})
}

func (r *RedisCacheStorage) FindByIDs(ctx context.Context, T resources.T, ids ...interface{}) iterators.Interface {
	c := r.client.WithContext(ctx)

	pipe, sender := iterators.NewPipe()
	go func() {
		defer sender.Close()

		rT := reflect.TypeOf(T)
		for _, id := range ids {
			get := c.Get(r.getKey(T, id))
			if err := get.Err(); err != nil {
				if err == redis.Nil {
					sender.Error(fmt.Errorf(`%T is not found by %v`, T, id))
					return
				}
				sender.Error(err)
				return
			}

			bytes, err := get.Bytes()
			if err != nil {
				sender.Error(err)
				return
			}

			ptr := reflect.New(rT)
			if err := r.unmarshal(bytes, ptr.Interface()); err != nil {
				sender.Error(err)
				return
			}

			if err := sender.Encode(ptr.Elem().Interface()); err != nil {
				return
			}
		}

	}()

	return pipe
}

func (r *RedisCacheStorage) marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (r *RedisCacheStorage) unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

//func (r *RedisCacheStorage) marshal(v interface{}) ([]byte, error) {
//	buf := bytes.NewBuffer([]byte{})
//	encoder := gob.NewEncoder(buf)
//	err := encoder.Encode(v)
//	return buf.Bytes(), err
//}
//
//func (r *RedisCacheStorage) unmarshal(data []byte, v interface{}) error {
//	decoder := gob.NewDecoder(bytes.NewReader(data))
//	return decoder.Decode(v)
//}

func (r *RedisCacheStorage) getKeyPrefix(T interface{}) string {
	return r.EntityTypeNameMappingFunc(T)
}

func (r *RedisCacheStorage) getKey(T, id interface{}) string {
	return fmt.Sprintf(`%s#%s`, r.getKeyPrefix(T), id)
}

func (r *RedisCacheStorage) getID(ptr interface{}) interface{} {
	id, _ := resources.LookupID(ptr)
	return id
}
