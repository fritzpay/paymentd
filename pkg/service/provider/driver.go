package provider

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/provider/fritzpay"
	"github.com/gorilla/mux"
	"net/http"
)

const (
	driverFritzpay int64 = iota
)

type Driver interface {
	Attach(ctx *service.Context, mux *mux.Router) error

	InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error)
}

var drivers map[int64]Driver

func init() {
	drivers = make(map[int64]Driver)
	drivers[driverFritzpay] = &fritzpay.Driver{}
}
