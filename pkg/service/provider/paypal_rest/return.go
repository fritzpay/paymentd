package paypal_rest

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"gopkg.in/inconshreveable/log15.v2"

	"net/http"
)

func (d *Driver) ReturnHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "ReturnHandler"})
		paymentIDStr := r.URL.Query().Get(paymentIDParam)
		if paymentIDStr == "" {
			log.Info("request without payment ID")
			d.NotFoundHandler(nil).ServeHTTP(w, r)
			return
		}
		nonce := r.URL.Query().Get(nonceParam)
		if nonce == "" {
			log.Info("request without nonce")
			d.NotFoundHandler(nil).ServeHTTP(w, r)
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
		payerID := r.URL.Query().Get(paypalPayerIDParameter)

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
		_, err = TransactionByPaymentIDAndNonceTx(tx, paymentID, nonce)
		if err != nil {
			if err == ErrTransactionNotFound {
				log.Info("paypal transaction not found")
				d.NotFoundHandler(nil).ServeHTTP(w, r)
				return
			}
			log.Error("error retrieving paypal transaction")
			d.InternalErrorHandler(nil).ServeHTTP(w, r)
			return
		}
		p, err := payment.PaymentByIDTx(tx, paymentID)
		if err != nil {
			if err == payment.ErrPaymentNotFound {
				log.Info("payment not found", log15.Ctx{"err": err})
				d.NotFoundHandler(nil).ServeHTTP(w, r)
				return
			}
			log.Error("error retrieving payment", log15.Ctx{"err": err})
			d.InternalErrorHandler(nil).ServeHTTP(w, r)
			return
		}
		currentTx, err := TransactionCurrentByPaymentIDTx(tx, p.PaymentID())
		if err != nil {
			if err == ErrTransactionNotFound {
				log.Crit("no transaction")
				d.InternalErrorHandler(p).ServeHTTP(w, r)
				return
			}
			log.Error("error retrieving current transaction", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}
		if currentTx.Type != TransactionTypeCreatePaymentResponse {
			log.Info("no execute payment required. skipping...")
			d.ReturnPageHandler(p).ServeHTTP(w, r)
			return
		}

		exec := &PayPalPaymentExecution{
			PayerID: payerID,
		}
		execJSON, err := json.Marshal(exec)
		if err != nil {
			log.Error("error encoding execute payment request", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}

		if currentTx.Links == nil {
			log.Crit("create payment response without links")
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}
		execTx := &Transaction{
			ProjectID: p.ProjectID(),
			PaymentID: p.ID(),
			Timestamp: time.Now(),
			Type:      TransactionTypeExecutePayment,
			Links:     currentTx.Links,
			Data:      execJSON,
		}
		execTx.SetPayerID(payerID)
		if currentTx.PaypalID.Valid {
			currentTx.SetPaypalID(currentTx.PaypalID.String)
		}
		err = InsertTransactionTx(tx, execTx)
		if err != nil {
			log.Error("error saving execute payment transaction", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}

		commit = true
		err = tx.Commit()
		if err != nil {
			log.Crit("error on commit tx", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}

		go d.executePayment(p, payerID)
	})
}

func (d *Driver) executePayment(p *payment.Payment, payerID string) {
	log := d.log.New(log15.Ctx{
		"method":    "executePayment",
		"projectID": p.ProjectID(),
		"paymentID": p.ID(),
		"payerID":   payerID,
	})
	log.Debug("executing payment...")
}
