package web

import (
	"net/http"
	"sync"
)

type ResponseWriter struct {
	w http.ResponseWriter

	mu            sync.Mutex
	statusCode    int
	headerWritten bool
	written       int
}

func (r *ResponseWriter) Header() http.Header {
	return r.w.Header()
}

func (r *ResponseWriter) WriteHeader(s int) {
	r.mu.Lock()
	if r.headerWritten {
		r.mu.Unlock()
		return
	}
	r.statusCode = s
	r.headerWritten = true
	r.mu.Unlock()
	r.w.WriteHeader(s)
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	w, err := r.w.Write(b)
	r.mu.Lock()
	r.written += w
	r.mu.Unlock()
	return w, err
}
