package signalflight

import (
	"sync"
)

type call struct {
	wg    sync.WaitGroup
	value interface{}
	err   error
}

type Group struct {
	mu sync.Mutex
	m  map[string]*call
}

//只保证同一时刻，同一个key的DB访问线程数为1;
func (g *Group) Do(key string, fn func() (value interface{}, err error)) (interface{}, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call)
	}
	//已有任务在进行则等待
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.value, c.err
	}

	c := new(call)
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	c.value, c.err = fn()
	c.wg.Done()

	//删除任务
	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.value, c.err
}
