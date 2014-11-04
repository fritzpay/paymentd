package fritzpay

import (
	"fmt"
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

func (d *Driver) Attach(ctx *service.Context, mux *mux.Router) error {
	d.ctx = ctx
	mux.HandleFunc(FritzpayDriverPath+"/status", d.Status)
	mux.Handle(FritzpayDriverPath+"/payment", d.PaymentInfo())
	return nil
}

func (d *Driver) InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error) {
	return nil, nil
}

func (d *Driver) Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "FritzPay OK.")
}

func (d *Driver) PaymentInfo() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}
