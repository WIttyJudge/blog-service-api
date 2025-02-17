package cache

import (
	"time"

	"github.com/maypok86/otter"
)

var defaultCapacity = 1000

type Cache[K comparable, V any] struct {
	cacheName string
	cache     *otter.CacheWithVariableTTL[K, V]
}

func New[K comparable, V any](capacity int, cacheName string) *Cache[K, V] {
	if capacity <= 0 {
		capacity = defaultCapacity
	}

	cache, _ := otter.MustBuilder[K, V](capacity).WithVariableTTL().Build()

	c := &Cache[K, V]{
		cacheName: cacheName,
		cache:     &cache,
	}

	return c
}

func (c *Cache[K, V]) Get(key K) (V, bool) {
	value, ok := c.cache.Get(key)
	if !ok {
		cacheMetrics.WithLabelValues(c.cacheName, "miss").Inc()
		return value, ok
	}

	cacheMetrics.WithLabelValues(c.cacheName, "hit").Inc()
	return value, ok
}

func (c *Cache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.cache.Set(key, value, ttl)
}

func (c *Cache[K, V]) Delete(key K) {
	c.cache.Delete(key)
}
