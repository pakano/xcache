package lfu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	lfu := NewLFU(6)
	testCases := map[string]Value{
		"a": String("aa"),
		"b": String("bb"),
		"c": String("cc"),
	}
	for k, v := range testCases {
		lfu.Set(k, v)
	}

	lfu.Get("a")
	lfu.Get("b")
	lfu.queue.Print()

	lfu.Set("d", String("dd"))

	for k := range testCases {
		_, ok := lfu.Get(k)
		if k == "c" {
			assert.False(t, ok)
		} else {
			assert.True(t, ok)
		}
	}
}

func TestLenSize(t *testing.T) {
	lfu := NewLFU(6)
	testCases := map[string]Value{
		"a": String("aa"),
		"b": String("bb"),
		"c": String("cc"),
	}
	for k, v := range testCases {
		lfu.Set(k, v)
	}

	assert.Equal(t, lfu.Size(), 6)
	assert.Equal(t, lfu.Len(), 3)
	lfu.Del("a")
	assert.Equal(t, lfu.Size(), 4)
	assert.Equal(t, lfu.Len(), 2)
}
