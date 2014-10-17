package payment

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
)

// API represents the payment API in version 1.x
type API struct {
	ctx *service.Context
	log log15.Logger
}

// NewAPI creates a new payment API
func NewAPI(ctx *service.Context) *API {
	a := &API{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1/payment"}),
	}
	return a
}
