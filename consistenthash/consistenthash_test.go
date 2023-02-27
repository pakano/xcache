package consistenthash

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func Randstr(n int) string {
	var charset string = "0123456789abcdefghijklmnopqrstuvwxyz"
	rand.Seed(time.Now().UnixNano())
	var str = make([]byte, n)
	for j := 0; j < n; j++ {
		str[j] = charset[rand.Intn(len(charset))]
	}
	return string(str)
}

func TestConsistentHash1(t *testing.T) {
	m := New(3, nil)
	m.Add("node1", "node2", "node3")

	c := make(map[string]int)

	fmt.Println(m.keys)

	for i := 0; i < 100000; i++ {
		key := Randstr(4)
		node := m.Get(key)
		c[node]++
		//fmt.Println(node)
	}
	fmt.Println(c)
}

func TestConsistentHash2(t *testing.T) {
	m := New(3, func(b []byte) uint32 {
		i, _ := strconv.Atoi(string(b))
		return uint32(i)
	})

	//[1 2 3 11 12 13 21 22 23]
	m.Add("1", "2", "3")

	testcases := map[string]string{
		"11": "1",
		"2":  "2",
		"22": "2",
		"8":  "1",
		"25": "1",
		"7":  "1",
		"13": "3",
		"14": "1",
		"15": "1",
	}

	for k, v := range testcases {
		if m.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
}
