package lfu

import (
	"container/heap"
	"reflect"
	"testing"
)

type String string

func (s String) Len() int {
	return len(s)
}

func TestFix(t *testing.T) {
	q := &queue{}
	en1 := &entry{key: "a", value: String("aa"), weight: 1}
	en2 := &entry{key: "b", value: String("bb"), weight: 2}
	en3 := &entry{key: "c", value: String("cc"), weight: 3}
	heap.Push(q, en1)
	heap.Push(q, en2)
	heap.Push(q, en3)
	heap.Init(q)

	if !reflect.DeepEqual((*q)[0], en1) {
		t.Fatal("1")
	}
	q.Update(en1, en1.value, 10)
	q.Update(en2, en2.value, 100)
	if !reflect.DeepEqual((*q)[0], en3) {
		t.Fatal("1")
	}
}

func TestPop(t *testing.T) {
	q := &queue{}
	en1 := &entry{key: "a", value: String("aa"), weight: 1}
	en2 := &entry{key: "b", value: String("bb"), weight: 2}
	en3 := &entry{key: "c", value: String("cc"), weight: 3}
	heap.Push(q, en1)
	heap.Push(q, en2)
	heap.Push(q, en3)
	heap.Init(q)
	heap.Remove(q, 0)
	if !reflect.DeepEqual((*q)[0], en2) {
		t.Fatal("1")
	}
}
