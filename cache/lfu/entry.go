package lfu

import (
	"container/heap"
	"fmt"
	"xcache/cache"
)

type entry struct {
	index  int
	weight int
	key    string
	value  cache.Value
}

type queue []*entry

func (q queue) Len() int {
	return len(q)
}

func (q queue) Less(i, j int) bool {
	return q[i].weight < q[j].weight
}

func (q queue) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *queue) Push(x interface{}) {
	en := x.(*entry)
	en.index = q.Len()
	*q = append(*q, en)
}

func (q *queue) Pop() interface{} {
	l := q.Len()
	en := (*q)[l-1]
	*q = (*q)[0 : l-1]
	return en
}

//堆重排
func (q *queue) Update(en *entry, val cache.Value, weight int) {
	en.value = val
	en.weight = weight
	heap.Fix(q, en.index)
}

func (q queue) Print() {
	for i := range q {
		fmt.Println(q[i])
	}
	fmt.Println("---------------------------")
}
