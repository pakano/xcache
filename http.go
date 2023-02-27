package xcache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"xcache/consistenthash"
	"xcache/xcachepb"

	"github.com/golang/protobuf/proto"
)

const (
	defalutBasePath = "/_xcache/"
	defalutReplicas = 50
)

type HttpPool struct {
	self       string
	basePath   string
	mu         sync.Mutex
	peers      *consistenthash.Map
	httpGetter map[string]*httpGetter
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defalutBasePath,
	}
}

func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (h *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, h.basePath) {
		http.Error(w, "basepath is invalid", http.StatusInternalServerError)
		return
	}

	// fmt.Println(h.basePath)

	parts := strings.SplitN(r.URL.Path[len(h.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "path is invalid", http.StatusInternalServerError)
		return
	}

	//fmt.Println(parts)

	groupName := parts[0]
	keyName := parts[1]

	g := GetGroup(groupName)
	if g == nil {
		http.Error(w, "group not found", http.StatusNotFound)
		return
	}
	v, err := g.Get(keyName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := proto.Marshal(&xcachepb.Response{Value: v.ByteSlice()})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/octet-stream")
	w.Write(data)
}

func (p *HttpPool) SetPeers(peers ...string) {
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
func (p *HttpPool) PeerPicker(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		//
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

var _ PeerGetter = (*httpGetter)(nil)
