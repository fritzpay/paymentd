package service

import (
	"code.google.com/p/go.net/context"
	"net/http"
	"sync"
	"time"
)

const (
	// the default purge timeout for the request context
	// no request should take longer than 5 minutes
	requestContextTimeout = 5 * time.Minute
)

const (
	// ContextVarAuthKey is the name of the key under which the auth container
	// will be stored in request contexts
	ContextVarAuthKey = "Auth"
)

var (
	mutex           sync.RWMutex
	requestContexts = make(map[*http.Request]context.Context)
)

// RequestContextTimeout returns the timeout after a request context is purged
var RequestContextTimeout func() time.Duration = func() time.Duration {
	return requestContextTimeout
}

type key int

const reqKey key = 0

type reqContext struct {
	context.Context
	r *http.Request
}

func (r *reqContext) Value(key interface{}) interface{} {
	if key == reqKey {
		return r.r
	}
	return r.Context.Value(key)
}

// SetRequestContext sets a new context for a request
func SetRequestContext(r *http.Request, ctx *Context) {
	mutex.Lock()
	c := &reqContext{ctx, r}
	requestContexts[r], _ = context.WithTimeout(c, RequestContextTimeout())
	go cancelRequestContext(requestContexts[r])
	mutex.Unlock()
}

// RequestContext returns a request associated with the given request
func RequestContext(r *http.Request) context.Context {
	mutex.RLock()
	ctx := requestContexts[r]
	mutex.RUnlock()
	return ctx
}

// SetRequestContextVar associates a var with a request context
func SetRequestContextVar(r *http.Request, key, value interface{}) {
	mutex.Lock()
	ctx := requestContexts[r]
	if ctx == nil {
		mutex.Unlock()
		return
	}
	requestContexts[r] = context.WithValue(ctx, key, value)
	mutex.Unlock()
}

// ClearRequestContext removes the associated context for the given request
func ClearRequestContext(r *http.Request) {
	mutex.Lock()
	delete(requestContexts, r)
	mutex.Unlock()
}

func cancelRequestContext(ctx context.Context) {
	select {
	case <-ctx.Done():
		if r, ok := ctx.Value(reqKey).(*http.Request); ok {
			ClearRequestContext(r)
		}
	}
}
