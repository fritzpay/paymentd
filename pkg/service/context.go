package service

import (
	"code.google.com/p/go.net/context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/config"
	"gopkg.in/inconshreveable/log15.v2"
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
	return c, nil
}
