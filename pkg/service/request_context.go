package service

import (
	"code.google.com/p/go.net/context"
	"net/http"
	"sync"
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

func SetRequestContext(r *http.Request, ctx *Context) {
	mutex.Lock()
	requestContexts[r] = ctx.Context
	mutex.Unlock()
}

func RequestContext(r *http.Request) context.Context {
	mutex.RLock()
	ctx := requestContexts[r]
	mutex.RUnlock()
	return ctx
}

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

func Clear(r *http.Request) {
	mutex.Lock()
	delete(requestContexts, r)
	mutex.Unlock()
}
