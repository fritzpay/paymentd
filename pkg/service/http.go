package service

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

var (
	ErrTimedOut = errors.New("http Write: already timed out")
)

func TimeoutHandler(logFunc func(msg string, ctx ...interface{}), d time.Duration, h http.Handler) http.Handler {
	f := func() <-chan time.Time {
		return time.After(d)
	}
	return &timeoutHandler{log: logFunc, handler: h, timeout: f}
}

type timeoutHandler struct {
	log     func(msg string, ctx ...interface{})
	handler http.Handler
	timeout func() <-chan time.Time
}

func (h *timeoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	done := make(chan struct{})
	tw := &timeoutWriter{w: w}
	go func() {
		h.handler.ServeHTTP(tw, r)
		close(done)
	}()
	select {
	case <-done:
		return
	case <-h.timeout():
		tw.mu.Lock()
		if !tw.wroteHeader {
			tw.w.WriteHeader(http.StatusServiceUnavailable)
		}
		tw.timedOut = true
		tw.mu.Unlock()
		h.log("request timeout",
			"requestURL", r.URL.String(),
		)
	}
}

type timeoutWriter struct {
	w http.ResponseWriter

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
}

func (w *timeoutWriter) Header() http.Header {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.w.Header()
}

func (w *timeoutWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.wroteHeader = true
	if w.timedOut {
		return 0, ErrTimedOut
	}
	return w.w.Write(p)
}

func (w *timeoutWriter) WriteHeader(status int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.timedOut || w.wroteHeader {
		return
	}
	w.wroteHeader = true
	w.w.WriteHeader(status)
}
