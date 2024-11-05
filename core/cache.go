package core

import (
	"time"

	"github.com/elastic/go-freelru"
	"github.com/zeebo/xxh3"
)

type Cache[K comparable, V any] interface {
	freelru.Cache[K, V]
}

type CacheConfig[K comparable] struct {
	HashKey  freelru.HashKeyCallback[K]
	Lifetime time.Duration
}

func CacheHashKeyString(key string) uint32 {
	return uint32(xxh3.HashString(key))
}

func NewCache[K comparable, V any](config CacheConfig[K]) (freelru.Cache[K, V], error) {
	lru, err := freelru.New[K, V](8192, config.HashKey)
	if config.Lifetime != 0 {
		lru.SetLifetime(config.Lifetime)
	}
	return lru, err
}
