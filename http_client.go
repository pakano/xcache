package xcache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
	"xcache/consistenthash"
	"xcache/xcachepb"

	servicediscover "xcache/service_discover"

	"github.com/golang/protobuf/proto"
)

const (
	defalutBasePath = "/_xcache/"
	defalutReplicas = 50
)

type HttpClient struct {
	mu           sync.Mutex
	peers        *consistenthash.Map
	httpGetter   map[string]*httpGetter
	registry     *servicediscover.Registry
	registryAddr string
}

func NewHttpClient(replicas int, fn consistenthash.Hash, timeout time.Duration, addr string) *HttpClient {
	return &HttpClient{
		peers:        consistenthash.New(replicas, fn),
		httpGetter:   make(map[string]*httpGetter),
		registry:     servicediscover.New(timeout),
		registryAddr: addr,
	}
}

func (p *HttpClient) Run(discoveryCycle time.Duration) {
	go func() {
		if err := http.ListenAndServe(p.registryAddr, p.registry); err != nil {
			log.Fatalln(err)
		}
	}()
	go func() {
		c := servicediscover.NewGeeRegistryDiscovery(fmt.Sprintf("http://%s", p.registryAddr), discoveryCycle)
		ticker := time.NewTicker(discoveryCycle)
		for {
			<-ticker.C
			c.Refresh()
			p.SetPeers(c.Servers()...)
		}
	}()
}

func (p *HttpClient) SetPeers(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consistenthash.New(defalutReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetter = make(map[string]*httpGetter)
	for _, peer := range peers {
		p.httpGetter[peer] = &httpGetter{baseURL: peer + defalutBasePath}
	}
}

//实现PeerPicker接口
func (p *HttpClient) PeerPicker(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" {
		return p.httpGetter[peer], true
	}
	return nil, false
}

type httpGetter struct {
	baseURL string
}

//实现PeerGetter接口
func (h *httpGetter) Get(in *xcachepb.Request, out *xcachepb.Response) error {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(in.Group),
		url.QueryEscape(in.Key),
	)

	res, err := http.Get(u)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response body: %v", err)
	}

	if err = proto.Unmarshal(bytes, out); err != nil {
		return err
	}

	return nil
}
