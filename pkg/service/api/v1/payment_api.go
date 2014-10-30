package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/payment"
	"gopkg.in/inconshreveable/log15.v2"
)

// API represents the payment API in the version 1.x
type PaymentAPI struct {
	ctx *service.Context
	log log15.Logger

	paymentService *payment.Service
}

// NewAPI creates a new payment API
func NewPaymentAPI(ctx *service.Context) (*PaymentAPI, error) {
	p := &PaymentAPI{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1",
			"API": "PaymentAPI",
		}),
	}
	var err error
	p.paymentService, err = payment.NewService(ctx)
	if err != nil {
		return nil, err
	}
	return p, nil
}
