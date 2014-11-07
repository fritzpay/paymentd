package notification

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/service"
)

// Callbacker describes a type that can provide information about callbacks to be made
type Callbacker interface {
	HasCallback() bool
	CallbackConfig() (url, apiVersion, projectKey string)
}

func CanCallback(c Callbacker) bool {
	return c.HasCallback()
}

func Notify(ctx *service.Context, c Callbacker, paymentTx *payment.PaymentTransaction) {

}
