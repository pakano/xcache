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

/*
goos: linux
goarch: amd64
pkg: xcache/consistenthash
cpu: Intel(R) Xeon(R) CPU E5-2620 v3 @ 2.40GHz
BenchmarkXxx-12         18219883                65.70 ns/op            0 B/op          0 allocs/op
PASS
ok      xcache/consistenthash   1.275s
*/
func BenchmarkXxx(b *testing.B) {
	m := New(3, nil)
	peers := []string{
		"http://127.0.0.1:8080",
		"http://127.0.0.1:8081",
		"http://127.0.0.1:8082",
	}
	m.Add(peers...)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		m.Get("1")
	}
}
