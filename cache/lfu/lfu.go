package lfu

import (
	"container/heap"
	"xcache/cache"
)

type LFU struct {
	maxBytes  int
	nBytes    int
	queue     *queue
	cache     map[string]*entry
	onEvicted cache.OnEvictedFunc
}

func NewLFU(maxBytes int, onEvicted cache.OnEvictedFunc) *LFU {
	queue := make(queue, 0)
	return &LFU{
		maxBytes:  maxBytes,
		queue:     &queue,
		cache:     make(map[string]*entry),
		onEvicted: onEvicted,
	}
}

func (l *LFU) Set(key string, val cache.Value) {
	if en, ok := l.cache[key]; ok {
		l.nBytes += val.Len() - en.value.Len()
		l.queue.Update(en, val, en.weight+1)
	} else {
		en := &entry{
			key:    key,
			value:  val,
			weight: 0,
		}

		heap.Push(l.queue, en)           // 插入queue 并重新排序为堆
		l.cache[key] = en                // 插入 map
		l.nBytes += len(key) + val.Len() // 更新内存占用

		// 如果超出内存长度，则删除最 '无用' 的元素，0表示无内存限制
		for l.maxBytes > 0 && l.nBytes > l.maxBytes {
			l.removeOldest()
		}
	}
}

// 获取指定元素,访问次数加1
func (l *LFU) Get(key string) (cache.Value, bool) {
	if en, ok := l.cache[key]; ok {
		l.queue.Update(en, en.value, en.weight+1)
		return en.value, true
	}
	return nil, false
}

// 删除指定元素（删除queue和map中的val）
func (l *LFU) Del(key string) {
	if en, ok := l.cache[key]; ok {
		heap.Remove(l.queue, en.index)
		l.removeElement(en)
	}
}

func (l *LFU) Len() int {
	return l.queue.Len()
}

func (l *LFU) removeOldest() {
	if l.queue.Len() == 0 {
		return
	}
	v := heap.Pop(l.queue)
	l.removeElement(v.(*entry))
}

func (l *LFU) removeElement(en *entry) {
	delete(l.cache, en.key)
	l.nBytes -= len(en.key) + en.value.Len()
	if l.onEvicted != nil {
		l.onEvicted(en.key, en.value)
	}
}
