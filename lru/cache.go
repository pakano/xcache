package lru

import "container/list"

type Cache struct {
	maxBytes  int64
	nBytes    int64
	cache     map[string]*list.Element
	ll        *list.List
	OnEvicted func(string, Value)
}

type Value interface {
	Len() int
}

type entry struct {
	key   string
	value Value
}

func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		cache:     make(map[string]*list.Element),
		ll:        list.New(),
		OnEvicted: onEvicted,
	}
}

func (c *Cache) removeOldest() {
	e := c.ll.Back()
	if e != nil {
		c.ll.Remove(e)
		en := e.Value.(*entry)
		delete(c.cache, en.key)
		c.nBytes -= int64(len(en.key)) + int64(en.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(en.key, en.value)
		}
	}
}

func (c *Cache) Set(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		en := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(en.value.Len())
		en.value = value
	} else {
		//头插
		en := &entry{key: key, value: value}
		ele := c.ll.PushFront(en)
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}

	//移除多余元素,maxBytes=0 表示无限多
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.removeOldest()
	}
}

func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		//先移动到首
		c.ll.MoveToFront(ele)
		en := ele.Value.(*entry)
		return en.value, true
	}
	return
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
