package paypal_rest

import (
	"code.google.com/p/godec/dec"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

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
		switch currentTx.Type {
		case TransactionTypeError:
			return d.PaymentErrorHandler(p), nil
		default:
			return d.InitPageHandler(p), nil
		}
	}

	cfg, err := ConfigByPaymentMethodTx(tx, method)
	if err != nil {
		log.Error("error retrieving PayPal config", log15.Ctx{"err": err})
		return nil, ErrDatabase
	}

	// create payment request
	req := &PayPalPaymentRequest{}
	if cfg.Type != "sale" && cfg.Type != "authorize" {
		log.Crit("invalid config type", log15.Ctx{"configType": cfg.Type})
		return nil, ErrInternal
	}
	req.Intent = cfg.Type
	req.Payer.PaymentMethod = PayPalPaymentMethodPayPal
	req.RedirectURLs, err = d.redirectURLs()
	if err != nil {
		log.Error("error creating redirect urls", log15.Ctx{"err": err})
		return nil, ErrInternal
	}
	req.Transactions = []PayPalTransaction{
		d.payPalTransactionFromPayment(p),
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
	paypalTx.Data.String, paypalTx.Data.Valid = string(jsonBytes), true

	err = InsertTransaction(tx, paypalTx)
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

	errors := make(chan error)
	go func() {
		for {
			select {
			case err := <-errors:
				if err == nil {
					return
				}
				log.Error("error on initializing", log15.Ctx{"err": err})
				return
			case <-d.ctx.Done():
				log.Warn("cancelled initialization", log15.Ctx{"err": d.ctx.Err()})
				return
			}
		}
	}()
	go d.doInit(errors, cfg, endpoint, p, string(jsonBytes))

	return d.InitPageHandler(p), nil
}

func (d *Driver) redirectURLs() (PayPalRedirectURLs, error) {
	u := PayPalRedirectURLs{}
	returnRoute, err := d.mux.Get("returnHandler").URLPath()
	if err != nil {
		return u, err
	}
	cancelRoute, err := d.mux.Get("cancelHandler").URLPath()
	if err != nil {
		return u, err
	}

	returnURL := &(*d.baseURL)
	returnURL.Path = returnRoute.Path
	u.ReturnURL = returnURL.String()

	cancelURL := &(*d.baseURL)
	cancelURL.Path = cancelRoute.Path
	u.CancelURL = cancelURL.String()

	return u, nil
}

func (d *Driver) payPalTransactionFromPayment(p *payment.Payment) PayPalTransaction {
	t := PayPalTransaction{}
	encPaymentID := d.paymentService.EncodedPaymentID(p.PaymentID())
	t.Custom = encPaymentID.String()
	t.InvoiceNumber = encPaymentID.String()
	amnt := &p.Decimal().Dec
	amnt.Round(amnt, dec.Scale(2), dec.RoundHalfUp)
	t.Amount = PayPalAmount{
		Currency: p.Currency,
		Total:    amnt.String(),
	}
	return t
}

func (d *Driver) doInit(errors chan<- error, cfg *Config, reqURL *url.URL, p *payment.Payment, body string) {
	log := d.log.New(log15.Ctx{
		"method":    "doPost",
		"projectID": p.ProjectID(),
		"paymentID": p.ID(),
		"methodKey": cfg.MethodKey,
	})
	if Debug {
		log.Debug("posting...")
	}

	tr, err := d.oAuthTransport(log)(p, cfg)
	if err != nil {
		log.Error("error on auth transport", log15.Ctx{"err": err})
		errors <- err
		return
	}
	err = tr.AuthenticateClient()
	if err != nil {
		log.Error("error authenticating", log15.Ctx{"err": err})
		errors <- err
		return
	}
	cl := tr.Client()
	resp, err := cl.Post(reqURL.String(), "application/json", strings.NewReader(body))
	if err != nil {
		log.Error("error on HTTP POST", log15.Ctx{"err": err})
		errors <- err
		return
	}
	if resp.StatusCode != http.StatusCreated {
		log.Error("error on HTTP request", log15.Ctx{"HTTPStatusCode": resp.StatusCode})
		errors <- ErrHTTP
		return
	}
	paypalP := &PaypalPayment{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(paypalP)
	if err != nil {
		log.Error("error decoding PayPal response", log15.Ctx{"err": err})
		errors <- ErrProvider
	}
	if Debug {
		log.Debug("received response", log15.Ctx{"paypalPayment": paypalP})
	}

	close(errors)
}
