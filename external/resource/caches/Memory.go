package caches

import (
	"context"
	"fmt"
	"reflect"

	"github.com/adamluzsi/frameless/iterators"
	"github.com/adamluzsi/frameless/reflects"
	"github.com/adamluzsi/frameless/resources"
	"github.com/adamluzsi/frameless/resources/storages/inmemory"
)

func NewInMemoryCacheStorage() *InMemoryCacheStorage {
	s := inmemory.New()
	s.Options.DisableEventLogging = true
	s.Options.DisableAsyncSubscriptionHandling = true
	s.Options.DisableRelativePathResolvingForTrace = true
	return &InMemoryCacheStorage{Storage: s}
}

type InMemoryCacheStorage struct {
	*inmemory.Storage
}

func (c InMemoryCacheStorage) Close() error {
	return nil
}

func (c *InMemoryCacheStorage) UpsertMany(ctx context.Context, ptrs ...interface{}) (rErr error) {
	tx, err := c.Storage.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if rErr != nil {
			_ = c.Storage.RollbackTx(tx)
		}
	}()

	for _, ptr := range ptrs {
		id, _ := resources.LookupID(ptr)
		n := reflect.New(reflects.BaseTypeOf(ptr)).Interface()
		if found, err := c.Storage.FindByID(ctx, n, id); err != nil {
			return err
		} else if found {
			if err := c.Storage.Update(tx, ptr); err != nil {
				return err
			}
			continue
		}

		if err := c.Storage.Create(tx, ptr); err != nil {
			return err
		}
	}
	return c.Storage.CommitTx(tx)
}

func (c *InMemoryCacheStorage) FindByIDs(ctx context.Context, T resources.T, ids ...interface{}) iterators.Interface {
	index := make(index)
	for _, id := range ids {
		index.Add(id)
	}

	pipe, sender := iterators.NewPipe()
	go func() {
		defer sender.Close()

		iter := iterators.Filter(c.Storage.FindAll(ctx, T), func(ent interface{}) bool {
			id, _ := resources.LookupID(ent)
			return index.Has(id)
		})

		var total int
		rT := reflect.TypeOf(T)
		for iter.Next() {
			total++
			ptr := reflect.New(rT)

			if err := iter.Decode(ptr.Interface()); err != nil {
				sender.Error(err)
				return
			}

			if err := sender.Encode(ptr.Elem().Interface()); err != nil {
				return
			}
		}

		if total != len(index) {
			sender.Error(fmt.Errorf(`not all %T was found by id: %d/%d`, T, total, len(index)))
		}
	}()
	return pipe
}
