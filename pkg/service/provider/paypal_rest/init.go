package paypal_rest

import (
	"net/http"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
)

func (d *Driver) InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error) {
	return nil, nil
}
