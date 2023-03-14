package cache

type CacheType int

const (
	LRUCACHE CacheType = iota
	LFUCACHE
	FIFOCACHE
)

//Value is an interface
type Value interface {
	Len() int
}

type OnEvictedFunc func(string, Value)

type Cache interface {
	Set(key string, value Value)
	Get(key string) (Value, bool)
	Del(key string)
	Len() int
}
