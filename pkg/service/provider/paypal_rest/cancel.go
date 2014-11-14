package paypal_rest

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

func (d *Driver) CancelHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "CancelHandler"})
		paymentIDStr := r.URL.Query().Get(paymentIDParam)
		if paymentIDStr == "" {
			log.Info("request without payment ID")
			d.NotFoundHandler().ServeHTTP(w, r)
			return
		}
		paymentID, err := payment.ParsePaymentIDStr(paymentIDStr)
		if err != nil {
			log.Warn("error parsing payment ID", log15.Ctx{
				"err":          err,
				"paymentIDStr": paymentIDStr,
			})
			d.BadRequestHandler().ServeHTTP(w, r)
			return
		}
		paymentID = d.paymentService.DecodedPaymentID(paymentID)
	})
}
