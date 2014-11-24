package service

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/fritzpay/paymentd/pkg/config"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	// ContextVarAuthKey is the name of the key under which the auth container
	// will be stored in request contexts
	ContextVarAuthKey = "Auth"
)

// Context is a custom context which is used by the service pkg
type Context struct {
	context.Context

	cfg config.Config
	log log15.Logger

	apiKeychain *Keychain
	webKeychain *Keychain

	principalDBWrite    *sql.DB
	principalDBReadOnly *sql.DB

	paymentDBWrite    *sql.DB
	paymentDBReadOnly *sql.DB

	rateLimit chan struct{}
}

// Value wraps the Context.Value
func (ctx *Context) Value(key interface{}) interface{} {
	switch key {
	case "cfg":
		return ctx.cfg
	case "log":
		return ctx.log
	case "keychain":
		return ctx.apiKeychain
	case "paymentDB":
		return ctx.paymentDBWrite
	case "principalDB":
		return ctx.principalDBWrite
	default:
		return ctx.Context.Value(key)
	}
}

// SetValue creates a new service context with the given value
func (ctx *Context) WithValue(key, value interface{}) *Context {
	return &Context{
		Context:             context.WithValue(ctx.Context, key, value),
		cfg:                 ctx.cfg,
		log:                 ctx.log,
		apiKeychain:         ctx.apiKeychain,
		webKeychain:         ctx.webKeychain,
		principalDBWrite:    ctx.principalDBWrite,
		principalDBReadOnly: ctx.principalDBReadOnly,
		paymentDBWrite:      ctx.paymentDBWrite,
		paymentDBReadOnly:   ctx.paymentDBReadOnly,
		rateLimit:           ctx.rateLimit,
	}
}

// Config returns the config.Config associated with the context
func (ctx *Context) Config() *config.Config {
	return &ctx.cfg
}

// Log returns the log15.Logger associated with the context
func (ctx *Context) Log() log15.Logger {
	return ctx.log
}

// Keychain returns the authorization container keychain associated with the context
func (ctx *Context) APIKeychain() *Keychain {
	return ctx.apiKeychain
}

func (ctx *Context) WebKeychain() *Keychain {
	return ctx.webKeychain
}

type dbRequestReadOnly bool

// ReadOnly is a possible parameter for the ctx.xDB() methods. If this parameter
// is passed to the methods, they will attempt to return the read-only database connection
var ReadOnly = dbRequestReadOnly(true)

// PrincipalDB returns the *sql.DB for the principal DB
// If the parameter(s) contain a service.ReadOnly, the read-only connection will be returned if present
func (ctx *Context) PrincipalDB(ros ...dbRequestReadOnly) *sql.DB {
	var ro bool
	if len(ros) > 0 {
		for _, r := range ros {
			if r {
				ro = true
			}
		}
	}
	if !ro {
		return ctx.principalDBWrite
	}
	if ctx.principalDBReadOnly == nil {
		return ctx.principalDBWrite
	}
	return ctx.principalDBReadOnly
}

// SetPrincipalDB sets the principal DB connection(s)
// It will panic if the write connection is nil
func (ctx *Context) SetPrincipalDB(w, ro *sql.DB) {
	if w == nil {
		panic("write DB connection cannot be nil")
	}
	ctx.principalDBWrite, ctx.principalDBReadOnly = w, ro
}

// PaymentDB returns the *sql.DB for the payment DB
// If the parameter(s) contain a service.ReadOnly, the read-only connection will be returned if present
func (ctx *Context) PaymentDB(ros ...dbRequestReadOnly) *sql.DB {
	var ro bool
	if len(ros) > 0 {
		for _, r := range ros {
			if r {
				ro = true
			}
		}
	}
	if !ro {
		return ctx.paymentDBWrite
	}
	if ctx.paymentDBReadOnly == nil {
		return ctx.paymentDBWrite
	}
	return ctx.paymentDBReadOnly
}

// SetPaymentDB sets the payment DB connection(s)
// It will panic if the write connection is nil
func (ctx *Context) SetPaymentDB(w, ro *sql.DB) {
	if w == nil {
		panic("write DB connection cannot be nil")
	}
	ctx.paymentDBWrite, ctx.paymentDBReadOnly = w, ro
}

func (ctx *Context) registerKeychain(kc *Keychain, keys []string) error {
	var err error
	for _, k := range keys {
		err = kc.AddKey(k)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctx *Context) registerKeychainFromConfig() error {
	err := ctx.registerKeychain(ctx.apiKeychain, ctx.Config().API.AuthKeys)
	if err != nil {
		return err
	}
	err = ctx.registerKeychain(ctx.webKeychain, ctx.Config().Web.AuthKeys)
	if err != nil {
		return err
	}
	return nil
}

func (ctx *Context) RateLimitHandler(parent http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-ctx.rateLimit
		defer func() {
			ctx.rateLimit <- struct{}{}
		}()
		parent.ServeHTTP(w, r)
	})
}

// NewContext creates a new service context for use in the service pkg
func NewContext(ctx context.Context, cfg config.Config, log log15.Logger) (*Context, error) {
	if log == nil {
		return nil, errors.New("log cannot be nil")
	}
	c := &Context{
		Context:     ctx,
		cfg:         cfg,
		log:         log,
		apiKeychain: NewKeychain(),
		webKeychain: NewKeychain(),
	}
	err := c.registerKeychainFromConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading keys from config: %v", err)
	}
	if cfg.Database.MaxOpenConns <= 0 {
		return nil, fmt.Errorf("invalid value for max open db conns %d", cfg.Database.MaxOpenConns)
	}
	c.rateLimit = make(chan struct{}, cfg.Database.MaxOpenConns)
	for i := 0; i < cfg.Database.MaxOpenConns; i++ {
		c.rateLimit <- struct{}{}
	}
	return c, nil
}

var (
	mutex           sync.RWMutex
	requestContexts = make(map[*http.Request]context.Context)
)

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
	requestContexts[r] = &reqContext{ctx, r}
	mutex.Unlock()
}

// RequestContext returns a request associated with the given request
func RequestContext(r *http.Request) context.Context {
	mutex.RLock()
	ctx := requestContexts[r]
	mutex.RUnlock()
	return ctx
}

// RequestAuthUserKey returns a request associated with the given request
func RequestContextAuth(r *http.Request) map[string]interface{} {
	mutex.RLock()
	ctx := requestContexts[r]
	mutex.RUnlock()
	return ctx.Value(ContextVarAuthKey).(map[string]interface{})
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
