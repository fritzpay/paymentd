package paypal_rest

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"

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
		tx, err = d.ctx.PaymentDB().Begin()
		if err != nil {
			commit = true
			log.Crit("error on begin tx", log15.Ctx{"err": err})
			d.InternalErrorHandler(nil).ServeHTTP(w, r)
			return
		}
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
		method, err := payment_method.PaymentMethodByIDTx(tx, p.Config.PaymentMethodID.Int64)
		if err != nil {
			log.Error("error retrieving payment method", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}
		if !method.Active() {
			log.Error("inactive payment method", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}
		cfg, err := ConfigByPaymentMethodTx(tx, method)
		if err != nil {
			log.Error("error retrieving config", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}
		currentTx, err := TransactionCurrentByPaymentIDTx(tx, p.PaymentID())
		if err != nil {
			log.Error("error retrieving current transaction", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}
		if currentTx.Type != TransactionTypeCreatePaymentResponse {
			if Debug {
				log.Debug("no execute payment required. skipping...")
			}
			d.statusHandler(currentTx, p, d.ReturnPageHandler(p)).ServeHTTP(w, r)
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
		links, err := currentTx.PayPalLinks()
		if err != nil {
			log.Error("error retrieving links", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}
		if links["execute"] == nil {
			log.Error("no execute link", log15.Ctx{"links": links})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}
		execURL, err := url.Parse(links["execute"].HRef)
		if err != nil {
			log.Error("error parsing execute URL", log15.Ctx{"err": err, "execURL": links["execute"].HRef})
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

		go d.executePayment(cfg, execURL, p, string(execJSON))

		d.statusHandler(execTx, p, d.ReturnPageHandler(p)).ServeHTTP(w, r)
	})
}

func (d *Driver) executePayment(cfg *Config, reqURL *url.URL, p *payment.Payment, body string) {
	log := d.log.New(log15.Ctx{
		"method":    "executePayment",
		"projectID": p.ProjectID(),
		"paymentID": p.ID(),
		"body":      body,
	})
	log.Debug("executing payment...")

	paymentTx, commitIntent, err := d.paymentService.IntentPaid(p, 500*time.Millisecond)
	if err != nil {
		log.Error("error on intent paid", log15.Ctx{"err": err})
		d.setPayPalError(p, nil)
		return
	}

	req, err := http.NewRequest("POST", reqURL.String(), strings.NewReader(body))
	if err != nil {
		log.Error("error creating execute payment request", log15.Ctx{"err": err})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	responseFunc := func(resp *http.Response, err error) error {
		if err != nil {
			log.Error("error on request", log15.Ctx{"err": err})
			return err
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Error("error reading response body", log15.Ctx{"err": err})
			d.setPayPalError(p, nil)
			return ErrHTTP
		}
		log = log.New(log15.Ctx{"responseBody": string(respBody)})
		if resp.StatusCode != http.StatusOK {
			log.Error("invalid HTTP status code", log15.Ctx{"statusCode": resp.StatusCode})
			d.setPayPalError(p, respBody)
			return ErrHTTP
		}
		if Debug {
			log.Debug("received response")
		}
		pay := &PaypalPayment{}
		err = json.Unmarshal(respBody, pay)
		if err != nil {
			log.Error("error decoding response", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrHTTP
		}
		paypalTx, err := NewPayPalPaymentTransaction(pay)
		if err != nil && paypalTx == nil {
			log.Error("error creating response transaction", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrInternal
		}
		if err != nil {
			log.Warn("error parsing response", log15.Ctx{"err": err})
		}
		paypalTx.ProjectID = p.ProjectID()
		paypalTx.PaymentID = p.ID()
		paypalTx.Type = TransactionTypeExecutePaymentResponse

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
			log.Crit("error on begin tx", log15.Ctx{"err": err})
			return ErrDatabase
		}

		err = InsertTransactionTx(tx, paypalTx)
		if err != nil {
			log.Error("error saving paypal transaction", log15.Ctx{"err": err})
			return ErrDatabase
		}

		paymentTx.Comment.String, paymentTx.Comment.Valid = "PayPal PaymentID: "+pay.ID, true
		err = d.paymentService.SetPaymentTransaction(tx, paymentTx)
		if err != nil {
			log.Error("error on payment transaction", log15.Ctx{"err": err})
			return ErrDatabase
		}

		commit = true
		err = tx.Commit()
		if err != nil {
			log.Crit("error on commit", log15.Ctx{"err": err})
			return ErrDatabase
		}
		commitIntent()

		return nil
	}
	err = httpDo(d.ctx, d.oAuthTransportFunc(p, cfg), req, responseFunc)
	if err != nil {
		log.Error("error on executing HTTP request", log15.Ctx{"err": err})
	}

}
