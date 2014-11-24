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
	tw := &timeoutWriter{ResponseWriter: w}
	go func() {
		h.handler.ServeHTTP(w, r)
		close(done)
	}()
	select {
	case <-done:
		return
	case <-h.timeout():
		tw.mu.Lock()
		if !tw.wroteHeader {
			tw.WriteHeader(http.StatusServiceUnavailable)
		}
		tw.timedOut = true
		tw.mu.Unlock()
		h.log("request timeout",
			"requestURL", r.URL.String(),
		)
	}
}

type timeoutWriter struct {
	http.ResponseWriter

	mu          sync.Mutex
	timedOut    bool
	wroteHeader bool
}

func (w *timeoutWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	timedOut := w.timedOut
	w.mu.Unlock()
	if timedOut {
		return 0, ErrTimedOut
	}
	return w.ResponseWriter.Write(p)
}

func (w *timeoutWriter) WriteHeader(status int) {
	w.mu.Lock()
	if w.timedOut || w.wroteHeader {
		w.mu.Unlock()
		return
	}
	w.wroteHeader = true
	w.mu.Unlock()
	w.ResponseWriter.WriteHeader(status)
}
