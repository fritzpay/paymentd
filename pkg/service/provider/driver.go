package provider

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	driverFritzpay int64 = 1
)

type Driver interface {
	Attach(ctx *service.Context, mux *mux.Router) error

	InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error)
}
