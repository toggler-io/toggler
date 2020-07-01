package caches

import "context"

type noCacheCtxKey struct{}

type noCache struct {
	done bool
}

func contextWithNoCache(ctx context.Context) context.Context {
	return context.WithValue(ctx, noCacheCtxKey{}, &noCache{})
}

func shouldSkipCache(ctx context.Context) bool {
	tx, ok := ctx.Value(noCacheCtxKey{}).(*noCache)
	if !ok {
		return false
	}

	return !tx.done
}

func noCacheDone(ctx context.Context) {
	if tx, ok := ctx.Value(noCacheCtxKey{}).(*noCache); ok {
		tx.done = true
	}
}
