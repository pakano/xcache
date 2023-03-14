package xcache

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"xcache/cache"
	"xcache/signalflight"
	"xcache/xcachepb"
)

//从数据源拉取数据
type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name       string
	getter     Getter
	maincache  *Cache
	peerPicker PeerPicker
	loader     *signalflight.Group //保护数据源
}

var (
	mu     sync.RWMutex //保护groups
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheType cache.CacheType, cacheBytes int, getter Getter) *Group {
	if getter == nil {
		panic("getter is nil")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		maincache: NewCache(cacheType, cacheBytes, nil),
		loader:    new(signalflight.Group),
	}

	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()

	return groups[name]
}

func (g *Group) Get(key string) (value ByteView, err error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}

	if v, ok := g.maincache.get(key); ok {
		//fmt.Println("cache hit")
		return v, nil
	}
	return g.load(key)
}

//从其他结点获取/从本地获取
func (g *Group) load(key string) (value ByteView, err error) {
	i, err := g.loader.Do(key, func() (interface{}, error) {
		//从远程获取
		if g.peerPicker != nil {
			if peer, ok := g.peerPicker.PeerPicker(key); ok {
				value, err := g.getFromPeer(peer, key)
				if err == nil {
					return value, nil
				}
				log.Println("[XCache] Failed to get from peer", err)
			} else if peer == nil {
				fmt.Println("from self")
			}
		}
		//如果远程结点获取失败，则从本地获取
		return g.getLocally(key)
	})
	if err == nil {
		return i.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (value ByteView, err error) {
	data, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value = ByteView{b: cloneBytes(data)}
	g.populateCache(key, value)
	return value, nil
}

//更新缓存
func (g *Group) populateCache(key string, value ByteView) {
	g.maincache.set(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if peers == nil {
		panic("peers cann't be nil")
	}
	g.peerPicker = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	in := xcachepb.Request{Group: g.name, Key: key}
	out := new(xcachepb.Response)
	err := peer.Get(&in, out)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: out.GetValue()}, nil
}
