// +build !go1.3

package paypal_rest

import (
	"bytes"
)

const (
	bufferSize = 2048
	buffers    = 16
)

var (
	bs = make(chan *bytes.Buffer, buffers)
)

func buffer() *bytes.Buffer {
	select {
	case buf := <-bs:
		buf.Reset()
		return buf
	default:
		return bytes.NewBuffer(make([]byte, 0, bufferSize))
	}
}

func putBuffer(buf *bytes.Buffer) {
	select {
	case bs <- buf:
	default:
	}
}
