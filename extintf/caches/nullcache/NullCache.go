package nullcache

import "github.com/adamluzsi/toggler/usecases"

func NewNullCache(s usecases.Storage) *NullCache {
	return &NullCache{Storage: s}
}

type NullCache struct{ usecases.Storage }
