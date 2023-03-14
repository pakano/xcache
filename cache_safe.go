package xcache

import (
	"sync"
	"xcache/cache"
	"xcache/cache/fifo"
	"xcache/cache/lfu"
	"xcache/cache/lru"
)

type Cache struct {
	mu    sync.Mutex
	cache cache.Cache
}

func NewCache(cacheType cache.CacheType, maxBytes int, onEvicted func(string, cache.Value)) (c *Cache) {
	c = new(Cache)

	switch cacheType {
	case cache.LRUCACHE:
		c.cache = lru.NewLRU(maxBytes, onEvicted)
	case cache.LFUCACHE:
		c.cache = lfu.NewLFU(maxBytes, onEvicted)
	case cache.FIFOCACHE:
		c.cache = fifo.NewFIFO(maxBytes, onEvicted)
	default:
		c.cache = lru.NewLRU(maxBytes, onEvicted)
	}
	return
}

func (c *Cache) set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache.Set(key, value)
}

func (c *Cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if v, ok := c.cache.Get(key); ok {
		return v.(ByteView), true
	}

	return
}
