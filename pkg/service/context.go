package service

import (
	"code.google.com/p/go.net/context"
	"database/sql"
	"errors"
	"github.com/fritzpay/paymentd/pkg/config"
	"gopkg.in/inconshreveable/log15.v2"
)

// Context is a custom context which is used by the service pkg
type Context struct {
	context.Context

	cfg config.Config
	log log15.Logger

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
	default:
		return ctx.Context.Value(key)
	}
}

// Config returns the config.Config associated with the context
func (ctx *Context) Config() config.Config {
	return ctx.cfg
}

// Log returns the log15.Logger associated with the context
func (ctx *Context) Log() log15.Logger {
	return ctx.log
}

// PrincipalDB returns the *sql.DB for the principal DB
// If the single parameter is true, the read-only connection will be returned if present
func (ctx *Context) PrincipalDB(ro bool) *sql.DB {
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
// If the single parameter is true, the read-only connection will be returned if present
func (ctx *Context) PaymentDB(ro bool) *sql.DB {
	if !ro {
		return ctx.paymentDBWrite
	}
	if ctx.paymentDBReadOnly == nil {
		return ctx.paymentDBWrite
	}
	return ctx.paymentDBReadOnly
}

// SetPymentDB sets the payment DB connection(s)
// It will panic if the write connection is nil
func (ctx *Context) SetPaymentDB(w, ro *sql.DB) {
	if w == nil {
		panic("write DB connection cannot be nil")
	}
	ctx.paymentDBWrite, ctx.paymentDBReadOnly = w, ro
}

// NewContext creates a new service context for use in the service pkg
func NewContext(ctx context.Context, cfg config.Config, log log15.Logger) (*Context, error) {
	if log == nil {
		return nil, errors.New("log cannot be nil")
	}
	return &Context{
		Context: ctx,
		cfg:     cfg,
		log:     log,
	}, nil
}
