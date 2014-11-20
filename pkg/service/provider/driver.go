package provider

import (
	"net/http"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
)

// Provider Driver Registry
//
// These names should match the provider names in the provider table
const (
	driverFritzpay   = "fritzpay"
	driverPaypalREST = "paypal_rest"
)

type Driver interface {
	Attach(ctx *service.Context, mux *mux.Router) error

	InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error)
}
