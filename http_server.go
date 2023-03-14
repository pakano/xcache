package xcache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	servicediscover "xcache/service_discover"
	"xcache/xcachepb"

	"github.com/golang/protobuf/proto"
)

type HttpPool struct {
	self     string
	basePath string
}

func NewHttpPool(self string, registryAddr string) *HttpPool {
	servicediscover.Heartbeat(fmt.Sprintf("http://%s", registryAddr), fmt.Sprintf("http://%s", self), 27*time.Second)
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

	//fmt.Println(r.URL.Path)

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
