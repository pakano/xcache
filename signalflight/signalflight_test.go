package signalflight

import (
	"fmt"
	"testing"
	"time"
)

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
