package xcache

import "xcache/xcachepb"

type PeerPicker interface {
	PeerPicker(key string) (PeerGetter, bool)
}

type PeerGetter interface {
	Get(in *xcachepb.Request, out *xcachepb.Response) error
}
