package fritzpay

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	FritzpayDriverPath = "/fritzpay"
)

type Driver struct {
	ctx *service.Context
}

func (d *Driver) Attach(ctx *service.Context, mux *mux.Router) {
	d.ctx = ctx
	mux.Handle(FritzpayDriverPath+"/payment", d.PaymentInfo())
}

func (d *Driver) InitPayment(p *payment.Payment, method *payment_method.PaymentMethod) (http.Handler, error) {
	return nil, nil
}

func (d *Driver) PaymentInfo() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}
