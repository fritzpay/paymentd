package web

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"time"
)

const (
	PaymentTokenParam = "token"
	PaymentCookieName = "payment"
)

func (h *Handler) authenticatePaymentRequest(w http.ResponseWriter, r *http.Request) (proceed bool) {
	log := h.log.New(log15.Ctx{"method": "authenticatePaymentRequest"})

	if tokenStr := r.URL.Query().Get(PaymentTokenParam); tokenStr != "" {
		var tx *sql.Tx
		var commit bool
		var err error
		defer func() {
			if tx != nil && !commit {
				err = tx.Rollback()
				if err != nil {
					log.Crit("error on rollback", log15.Ctx{"err": err})
				}
			}
		}()
		maxRetries := h.ctx.Config().Database.TransactionMaxRetries
		var retries int
	beginTx:
		if retries >= maxRetries {
			// no need to roll back
			commit = true
			log.Crit("too many retries on tx. aborting...", log15.Ctx{"maxRetries": maxRetries})
			w.WriteHeader(http.StatusInternalServerError)
			return false
		}
		tx, err = h.ctx.PaymentDB(service.ReadOnly).Begin()
		if err != nil {
			commit = true
			log.Crit("error on begin tx", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return false
		}
		p, err := h.paymentService.PaymentByToken(tx, tokenStr)
		if err != nil {
			if err == payment.ErrPaymentNotFound {
				w.WriteHeader(http.StatusNotFound)
				return false
			}
			log.Error("error retrieving payment token", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return false
		}
		if !p.Valid() {
			log.Crit("received invalid payment")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = h.paymentService.DeletePaymentToken(tx, tokenStr)
		if err != nil {
			if err == paymentService.ErrDBLockTimeout {
				retries++
				time.Sleep(time.Second)
				goto beginTx
			}
			log.Error("error deleting payment token", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	return true
}

func (h *Handler) PaymentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// log := h.log.New(log15.Ctx{"method": "PaymentHandler"})
		if !h.authenticatePaymentRequest(w, r) {
			return
		}
	})
}
