package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
)

// API represents the payment API in the version 1.x
type PaymentAPI struct {
	ctx *service.Context
	log log15.Logger
}

// NewAPI creates a new payment API
func NewPaymentAPI(ctx *service.Context) *PaymentAPI {
	p := &PaymentAPI{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1",
			"API": "PaymentAPI",
		}),
	}
	return p
}
