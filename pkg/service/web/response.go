package web

import (
	"net/http"
	"sync"
)

type ResponseWriter struct {
	http.ResponseWriter

	mu             sync.Mutex
	HTTPStatusCode int
	HeaderWritten  bool
	Written        int
}

func (r *ResponseWriter) WriteHeader(s int) {
	r.mu.Lock()
	if r.HeaderWritten {
		r.mu.Unlock()
		return
	}
	r.HTTPStatusCode = s
	r.HeaderWritten = true
	r.mu.Unlock()
	r.ResponseWriter.WriteHeader(s)
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	w, err := r.ResponseWriter.Write(b)
	r.mu.Lock()
	r.Written += w
	r.mu.Unlock()
	return w, err
}
