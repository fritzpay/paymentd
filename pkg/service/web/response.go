package web

import (
	"net/http"
)

type ResponseWriter struct {
	http.ResponseWriter

	HTTPStatusCode int
	HeaderWritten  bool
	Written        int
}

func (r *ResponseWriter) WriteHeader(s int) {
	r.HTTPStatusCode = s
	r.HeaderWritten = true
	r.ResponseWriter.WriteHeader(s)
}

func (r *ResponseWriter) Write(b []byte) (int, error) {
	w, err := r.ResponseWriter.Write(b)
	r.Written += w
	return w, err
}
