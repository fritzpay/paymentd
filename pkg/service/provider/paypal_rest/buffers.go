// +build go1.3

package paypal_rest

import (
	"bytes"
	"sync"
)

const (
	bufferSize = 2048
	buffers    = 16
)

var (
	bs = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, bufferSize))
		},
	}
)

func buffer() *bytes.Buffer {
	buf := bs.Get().(*bytes.Buffer)
	buf.Reset()
	return buf
}

func putBuffer(buf *bytes.Buffer) {
	bs.Put(buf)
}
