package admin

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
)

// API represents the admin API in version 1.x
type API struct {
	ctx *service.Context
	log log15.Logger
}

// NewAPI creates a new admin API
func NewAPI(ctx *service.Context) *API {
	a := &API{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1/admin"}),
	}
	return a
}
