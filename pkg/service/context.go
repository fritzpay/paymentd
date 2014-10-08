package service

import (
	"code.google.com/p/go.net/context"
	"github.com/fritzpay/paymentd/pkg/config"
	"gopkg.in/inconshreveable/log15.v2"
)

type serviceContext struct {
	context.Context
	cfg config.Config
	log log15.Logger
}

func (ctx *serviceContext) Value(key interface{}) interface{} {
	switch key {
	case "cfg":
		return ctx.cfg
	case "log":
		return ctx.log
	default:
		return ctx.Context.Value(key)
	}
}

func NewContext(ctx context.Context, cfg config.Config, log log15.Logger) context.Context {
	return &serviceContext{
		Context: ctx,
		cfg:     cfg,
		log:     log,
	}
}
