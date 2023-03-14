package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	"xcache"
	"xcache/cache"
	"xcache/xcachepb"
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

func startCacheServer(addr string, registryAddr string, g *xcache.Group) {
	pool := xcache.NewHttpPool(addr, registryAddr)
	log.Println("cache server is running at", addr)
	log.Fatalln(http.ListenAndServe(addr, pool))
}

func startRegistry(addr string) *xcache.HttpClient {
	timeout := time.Second * 30
	httpClient := xcache.NewHttpClient(3, nil, timeout, addr)
	httpClient.Run(timeout / 3)
	return httpClient
}

func startAPIServer(apiAddr string, httpclient *xcache.HttpClient) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		group := r.URL.Query().Get("group")
		key := r.URL.Query().Get("key")
		peerGetter, ok := httpclient.PeerPicker(key)
		if !ok {
			http.Error(w, "no cache server", http.StatusNotFound)
			return
		}

		in := xcachepb.Request{Group: group, Key: key}
		out := new(xcachepb.Response)
		err := peerGetter.Get(&in, out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(out.GetValue())
	})
	log.Println("api server is running in", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr, mux))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "XCache server port")
	flag.BoolVar(&api, "api", false, "Start a api server")
	flag.Parse()

	registryAddr := "192.168.4.41:9899"

	if api {
		httpclient := startRegistry(registryAddr)
		startAPIServer("192.168.4.41:9898", httpclient)
	} else {
		g := createGroup()
		startCacheServer(fmt.Sprintf("192.168.4.41:%d", port), registryAddr, g)
	}
}
