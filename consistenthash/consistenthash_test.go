package consistenthash

import (
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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

func TestConsistentHash(t *testing.T) {
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

func TestConsistentHashDel(t *testing.T) {
	m := New(3, func(b []byte) uint32 {
		i, _ := strconv.Atoi(string(b))
		return uint32(i)
	})

	m.Add("1", "2", "3")
	assert.Equal(t, m.keys, []uint32{1, 2, 3, 11, 12, 13, 21, 22, 23})

	m.Del("1")
	assert.Equal(t, m.keys, []uint32{2, 3, 12, 13, 22, 23})
}
