package core

import (
	"time"

	"github.com/elastic/go-freelru"
	"github.com/zeebo/xxh3"
)

type Cache[K comparable, V any] interface {
	freelru.Cache[K, V]
	GetName() string
}

type cacheLRU[K comparable, V any] struct {
	*freelru.LRU[K, V]
	name string
}

func (lru *cacheLRU[K, V]) GetName() string {
	return lru.name
}

type CacheConfig[K comparable] struct {
	HashKey  freelru.HashKeyCallback[K]
	Lifetime time.Duration
	Capacity uint32
	Name     string
}

func CacheHashKeyString(key string) uint32 {
	return uint32(xxh3.HashString(key))
}

func NewCache[K comparable, V any](config *CacheConfig[K]) Cache[K, V] {
	if config.Capacity == 0 {
		config.Capacity = 2048
	}
	lru, err := freelru.New[K, V](config.Capacity, config.HashKey)
	if err != nil {
		errMsg := "failed to create cache"
		if config.Name != "" {
			errMsg += ": " + config.Name
		}
		panic(errMsg)
	}
	if config.Lifetime != 0 {
		lru.SetLifetime(config.Lifetime)
	}
	cache := &cacheLRU[K, V]{LRU: lru, name: config.Name}
	return cache
}
