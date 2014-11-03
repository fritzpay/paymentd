package web

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"net/http"
)

func (h *Handler) SelectPaymentMethodHandler(p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
