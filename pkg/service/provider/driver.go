package provider

import (
	"net/http"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
)

const (
	driverFritzpay int64 = 1
)

type Driver interface {
	Attach(ctx *service.Context, mux *mux.Router) error

	InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error)
}
