package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func([]byte) uint32

type Map struct {
	hash     Hash
	replicas int
	keys     []uint32
	hashMap  map[uint32]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		keys:     nil,
		hashMap:  make(map[uint32]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			h := m.hash([]byte(strconv.Itoa(i) + key))
			m.keys = append(m.keys, h)
			m.hashMap[h] = key
		}
	}
	sort.Slice(m.keys, func(i, j int) bool {
		return m.keys[i] < m.keys[j]
	})
}

func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}

	h := m.hash([]byte(key))

	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= h
	})

	idx = idx % len(m.keys)

	//虚拟节点到实体结点
	return m.hashMap[m.keys[idx]]
}

func (m *Map) Del(key string) {
	if len(key) == 0 {
		return
	}

	for i := 0; i < m.replicas; i++ {
		h := m.hash([]byte(strconv.Itoa(i) + key))
		k := 0
		for j := range m.keys {
			if m.keys[j] != h {
				m.keys[k] = m.keys[j]
				k++
			}
		}
		m.keys = m.keys[:k]
		delete(m.hashMap, h)
	}
}
