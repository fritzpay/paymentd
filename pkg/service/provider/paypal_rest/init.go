package paypal_rest

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/nonce"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
)

const (
	paypalPaymentPath = "/v1/payments/payment"
)

func (d *Driver) InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error) {
	log := d.log.New(log15.Ctx{
		"method":          "InitPayment",
		"projectID":       p.ProjectID(),
		"paymentID":       p.ID(),
		"paymentMethodID": method.ID,
	})

	var tx *sql.Tx
	var err error
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
		return nil, ErrDatabase
	}

	currentTx, err := TransactionCurrentByPaymentIDTx(tx, p.PaymentID())
	if err != nil && err != ErrTransactionNotFound {
		log.Error("error retrieving transaction", log15.Ctx{"err": err})
		return nil, ErrDatabase
	}
	if err == nil {
		if Debug {
			log.Debug("already initialized payment")
		}
		return d.StatusHandler(currentTx, p), nil
	}

	cfg, err := ConfigByPaymentMethodTx(tx, method)
	if err != nil {
		log.Error("error retrieving PayPal config", log15.Ctx{"err": err})
		return nil, ErrDatabase
	}

	// create payment request
	non, err := nonce.New()
	if err != nil {
		log.Error("error generating nonce", log15.Ctx{"err": err})
		return nil, ErrInternal
	}
	req, err := d.createPaypalPaymentRequest(p, cfg, non)
	if err != nil {
		log.Error("error creating paypal payment request", log15.Ctx{"err": err})
		return nil, ErrInternal
	}
	if Debug {
		log.Debug("created paypal payment request", log15.Ctx{"request": req})
	}

	endpoint, err := url.Parse(cfg.Endpoint)
	if err != nil {
		log.Error("error on endpoint URL", log15.Ctx{"err": err})
		return nil, ErrInternal
	}
	endpoint.Path = paypalPaymentPath

	jsonBytes, err := json.Marshal(req)
	if err != nil {
		log.Error("error encoding request", log15.Ctx{"err": err})
		return nil, ErrInternal
	}

	paypalTx := &Transaction{
		ProjectID: p.ProjectID(),
		PaymentID: p.ID(),
		Timestamp: time.Now(),
		Type:      TransactionTypeCreatePayment,
	}
	paypalTx.SetIntent(cfg.Type)
	paypalTx.SetNonce(non.Nonce)
	paypalTx.Data = jsonBytes

	err = InsertTransactionTx(tx, paypalTx)
	if err != nil {
		log.Error("error saving transaction", log15.Ctx{"err": err})
		return nil, ErrDatabase
	}

	commit = true
	err = tx.Commit()
	if err != nil {
		log.Crit("error on commit", log15.Ctx{"err": err})
		return nil, ErrDatabase
	}

	go d.doInit(cfg, endpoint, p, string(jsonBytes))

	return d.InitPageHandler(p), nil
}

func (d *Driver) doInit(cfg *Config, reqURL *url.URL, p *payment.Payment, body string) {
	log := d.log.New(log15.Ctx{
		"method":      "doInit",
		"projectID":   p.ProjectID(),
		"paymentID":   p.ID(),
		"methodKey":   cfg.MethodKey,
		"requestBody": body,
	})
	if Debug {
		log.Debug("posting...")
	}

	req, err := http.NewRequest("POST", reqURL.String(), strings.NewReader(body))
	if err != nil {
		log.Error("error creating HTTP request", log15.Ctx{"err": err})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	responseFunc := func(resp *http.Response, err error) error {
		if err != nil {
			log.Error("error on HTTP", log15.Ctx{"err": err})
			d.setPayPalError(p, nil)
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
		if Debug {
			log.Debug("received response")
		}
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			log.Error("error on HTTP request", log15.Ctx{"HTTPStatusCode": resp.StatusCode})
			d.setPayPalError(p, respBody)
			return ErrHTTP
		}
		paypalP := &PaypalPayment{}
		err = json.Unmarshal(respBody, paypalP)
		if err != nil {
			log.Error("error decoding PayPal response", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrProvider
		}

		paypalTx := &Transaction{
			ProjectID: p.ProjectID(),
			PaymentID: p.ID(),
			Timestamp: time.Now(),
			Type:      TransactionTypeCreatePaymentResponse,
		}
		if paypalP.Intent != "" {
			paypalTx.SetIntent(paypalP.Intent)
		}
		if paypalP.ID != "" {
			paypalTx.SetPaypalID(paypalP.ID)
		}
		if paypalP.State != "" {
			paypalTx.SetState(paypalP.State)
		}
		if paypalP.CreateTime != "" {
			t, err := time.Parse(time.RFC3339, paypalP.CreateTime)
			if err != nil {
				log.Warn("error parsing paypal create time", log15.Ctx{"err": err})
			} else {
				paypalTx.PaypalCreateTime = &t
			}
		}
		if paypalP.UpdateTime != "" {
			t, err := time.Parse(time.RFC3339, paypalP.UpdateTime)
			if err != nil {
				log.Warn("error parsing paypal update time", log15.Ctx{"err": err})
			} else {
				paypalTx.PaypalUpdateTime = &t
			}
		}
		paypalTx.Links, err = json.Marshal(paypalP.Links)
		if err != nil {
			log.Error("error on saving links on response", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrProvider
		}
		paypalTx.Data, err = json.Marshal(paypalP)
		if err != nil {
			log.Error("error marshalling paypal payment response", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrProvider
		}
		err = InsertTransactionDB(d.ctx.PaymentDB(), paypalTx)
		if err != nil {
			log.Error("error saving paypal response", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrProvider
		}
		return nil
	}

	err = httpDo(d.ctx, d.oAuthTransportFunc(p, cfg), req, responseFunc)
	if err != nil {
		log.Error("error on create payment request", log15.Ctx{"err": err})
	}
}
