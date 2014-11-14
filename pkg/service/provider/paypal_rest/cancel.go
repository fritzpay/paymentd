package paypal_rest

import (
	"database/sql"
	"net/http"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"gopkg.in/inconshreveable/log15.v2"
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

		var tx *sql.Tx
		var commit bool
		defer func() {
			if tx != nil && !commit {
				err = tx.Rollback()
				if err != nil {
					log.Crit("error on rollback", log15.Ctx{"err": err})
				}
			}
		}()
		tx, err = d.ctx.PaymentDB().Begin()
		if err != nil {
			commit = true
			log.Crit("error on begin tx", log15.Ctx{"err": err})
			d.InternalErrorHandler(nil).ServeHTTP(w, r)
			return
		}

		p, err := payment.PaymentByIDTx(tx, paymentID)
		if err != nil {
			if err == payment.ErrPaymentNotFound {
				log.Info("payment not found")
				d.NotFoundHandler().ServeHTTP(w, r)
				return
			}
			log.Error("error retrieving payment", log15.Ctx{"err": err})
			d.InternalErrorHandler(nil).ServeHTTP(w, r)
			return
		}
		paypalTx, err := TransactionCurrentByPaymentIDTx(tx, p.PaymentID())
		if err != nil {
			if err == ErrTransactionNotFound {
				log.Info("paypal transaction not found")
				d.NotFoundHandler().ServeHTTP(w, r)
				return
			}
			log.Error("error retrievin paypal transaction")
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}
		if !paypalTx.PaypalID.Valid {
			log.Warn("paypal transaction without paypal id")
			d.PaymentErrorHandler(p).ServeHTTP(w, r)
			return
		}
	})
}
