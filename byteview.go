package xcache

type ByteView struct {
	b []byte
}

func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

func (v ByteView) Len() int {
	return len(v.b)
}

func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	data := make([]byte, len(b))
	copy(data, b)
	return data
}
