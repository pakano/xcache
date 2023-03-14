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
	"xcache/signalflight"
	"xcache/xcachepb"

	servicediscover "xcache/service_discover"

	"github.com/golang/protobuf/proto"
)

const (
	defalutBasePath = "/_xcache/"
	defalutReplicas = 50
)

type HttpClient struct {
	mu           sync.RWMutex
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
		c.Refresh()
		p.SetPeers(c.Servers()...)
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
		p.httpGetter[peer] = &httpGetter{baseURL: peer + defalutBasePath, loader: new(signalflight.Group)}
	}
}

//实现PeerPicker接口
func (p *HttpClient) PeerPicker(key string) (PeerGetter, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if peer := p.peers.Get(key); peer != "" {
		return p.httpGetter[peer], true
	}
	return nil, false
}

type httpGetter struct {
	baseURL string
	loader  *signalflight.Group
}

var client *http.Client

func init() {
	tr := http.DefaultTransport.(*http.Transport)
	tr2 := tr.Clone()
	tr2.MaxConnsPerHost = 20
	client = &http.Client{
		Timeout:   time.Second * 2,
		Transport: tr,
	}
}

//实现PeerGetter接口
func (h *httpGetter) Get(in *xcachepb.Request, out *xcachepb.Response) error {
	u := fmt.Sprintf(
		"%s%s/%s",
		h.baseURL,
		url.QueryEscape(in.Group),
		url.QueryEscape(in.Key),
	)

	//保护远程结点,提高性能
	v, err := h.loader.Do(in.Group+in.Key, func() (value interface{}, err error) {
		res, err := client.Get(u)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("server returned: %v", res.Status)
		}

		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body: %v", err)
		}
		if err != nil {
			return nil, fmt.Errorf("reading response body: %v", err)
		}

		if err = proto.Unmarshal(bytes, out); err != nil {
			return nil, err
		}
		return out, nil
	})
	if err != nil {
		return err
	}
	out = v.(*xcachepb.Response)

	return nil
}
