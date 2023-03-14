package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"xcache"
	"xcache/cache"
)

var db = map[string]string{
	"a": "aaa",
	"b": "bbb",
	"c": "ccc",
}

func createGroup() *xcache.Group {
	return xcache.NewGroup("demo", cache.LFUCACHE, 2<<10, xcache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if data, ok := db[key]; ok {
			return []byte(data), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func startCacheServer(addr string, peers []string, g *xcache.Group) {
	pool := xcache.NewHttpPool(addr)
	pool.SetPeers(peers...)
	g.RegisterPeers(pool)
	log.Println("xcache is running at", addr)

	log.Fatalln(http.ListenAndServe(addr[7:], pool))
}

func startAPIServer(apiAddr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		group := r.URL.Query().Get("group")
		key := r.URL.Query().Get("key")
		g := xcache.GetGroup(group)
		if g == nil {
			http.Error(w, "group not found", http.StatusNotFound)
			return
		}
		v, err := g.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(v.ByteSlice())
	})
	log.Println("api server is running in", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], mux))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "XCache server port")
	flag.BoolVar(&api, "api", false, "Start a api server")
	flag.Parse()

	apiAddr := "http://192.168.4.41:9999"
	addrMap := map[int]string{
		8001: "http://192.168.4.41:8001",
		8002: "http://192.168.4.41:8002",
		8003: "http://192.168.4.41:8003",
	}

	var addrs []string
	for _, addr := range addrMap {
		addrs = append(addrs, addr)
	}

	g := createGroup()
	if api {
		startAPIServer(apiAddr)
	} else {
		startCacheServer(addrMap[port], addrs, g)
	}
}
