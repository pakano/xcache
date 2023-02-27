package signalflight

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestXxx(t *testing.T) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	for i := 0; i < 10; i++ {
		go func() {
			wg.Wait()
			fmt.Println(1)
		}()
	}
	time.Sleep(time.Second)
	wg.Done()
	time.Sleep(time.Second)
}

func TestSignalFlight(t *testing.T) {
	g := new(Group)
	for i := 0; i < 10; i++ {
		go func() {
			g.Do("aa", func() (value interface{}, err error) {
				time.Sleep(time.Millisecond)
				fmt.Println(1)
				return
			})
		}()
	}
	time.Sleep(time.Second)
}
