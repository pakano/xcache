package lru

import (
	"container/list"
	"xcache/cache"
)

type LRU struct {
	maxBytes  int
	nBytes    int
	cache     map[string]*list.Element
	ll        *list.List
	onEvicted cache.OnEvictedFunc
}

type entry struct {
	key   string
	value cache.Value
}

func NewLRU(maxBytes int, onEvicted cache.OnEvictedFunc) *LRU {
	return &LRU{
		maxBytes:  maxBytes,
		cache:     make(map[string]*list.Element),
		ll:        list.New(),
		onEvicted: onEvicted,
	}
}

func (c *LRU) Set(key string, value cache.Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		en := ele.Value.(*entry)
		c.nBytes += value.Len() - en.value.Len()
		en.value = value
	} else {
		//头插
		en := &entry{key: key, value: value}
		ele := c.ll.PushFront(en)
		c.cache[key] = ele
		c.nBytes += len(key) + value.Len()
	}

	//移除多余元素,maxBytes=0 表示无限多
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.removeOldest()
	}
}

func (c *LRU) Get(key string) (value cache.Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		//先移动到首
		c.ll.MoveToFront(ele)
		en := ele.Value.(*entry)
		return en.value, true
	}
	return
}

func (c *LRU) Del(key string) {
	if ele, ok := c.cache[key]; ok {
		c.removeElement(ele)
	}
}

func (c *LRU) Len() int {
	return c.ll.Len()
}

func (c *LRU) removeOldest() {
	e := c.ll.Back()
	if e != nil {
		c.removeElement(e)
	}
}

func (c *LRU) removeElement(elem *list.Element) {
	c.ll.Remove(elem)
	en := elem.Value.(*entry)
	delete(c.cache, en.key)
	c.nBytes -= len(en.key) + en.value.Len()
	if c.onEvicted != nil {
		c.onEvicted(en.key, en.value)
	}
}
