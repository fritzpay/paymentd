package paypal_rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/nonce"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"

	"code.google.com/p/goauth2/oauth"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"

	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	// PaypalDriverPath is the (sub-)path under which PayPal driver endpoints
	// will be attached
	PaypalDriverPath = "/paypal"
)

const (
	providerTemplateDir = "paypal_rest"
	defaultLocale       = "en_US"
	// endpoint path for REST API URL
	paypalPaymentPath = "/v1/payments/payment"
)

var (
	ErrDatabase = errors.New("database error")
	ErrInternal = errors.New("paypal driver internal error")
	ErrHTTP     = errors.New("HTTP error")
	ErrProvider = errors.New("provider error")
)

// Driver is the PayPal provider driver
type Driver struct {
	ctx *service.Context
	mux *mux.Router
	log log15.Logger

	tmplDir string

	paymentService *paymentService.Service

	oauth *OAuthTransportStore
}

func (d *Driver) Attach(ctx *service.Context, mux *mux.Router) error {
	d.ctx = ctx
	d.log = ctx.Log().New(log15.Ctx{
		"pkg": "github.com/fritzpay/paymentd/pkg/service/provider/paypal_rest",
	})

	var err error
	d.paymentService, err = paymentService.NewService(ctx)
	if err != nil {
		d.log.Error("error initializing payment service", log15.Ctx{"err": err})
		return err
	}

	cfg := ctx.Config()
	if cfg.Provider.ProviderTemplateDir == "" {
		return fmt.Errorf("provider template dir not set")
	}
	d.tmplDir = path.Join(cfg.Provider.ProviderTemplateDir, providerTemplateDir)
	dirInfo, err := os.Stat(d.tmplDir)
	if err != nil {
		d.log.Error("error opening template dir", log15.Ctx{
			"err":     err,
			"tmplDir": d.tmplDir,
		})
		return err
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("provider template dir %s is not a directory", d.tmplDir)
	}
	_, err = url.Parse(cfg.Provider.URL)
	if err != nil {
		d.log.Error("error parsing provider base URL", log15.Ctx{"err": err})
		return fmt.Errorf("error on provider base URL: %v", err)
	}

	driverRoute := mux.PathPrefix(PaypalDriverPath)
	u, err := driverRoute.URLPath()
	if err != nil {
		d.log.Error("error determining path prefix", log15.Ctx{"err": err})
		return fmt.Errorf("error on subroute path: %v", err)
	}
	d.mux = driverRoute.Subrouter()
	d.mux.Handle("/return", ctx.RateLimitHandler(d.ReturnHandler())).Name("returnHandler")
	d.mux.Handle("/cancel", ctx.RateLimitHandler(d.CancelHandler())).Name("cancelHandler")
	staticDir := path.Join(d.tmplDir, "static")
	d.log.Info("serving static dir", log15.Ctx{
		"staticDir": staticDir,
		"prefix":    u.Path + "/static",
	})
	d.mux.PathPrefix("/static").Handler(http.StripPrefix(u.Path+"/static", http.FileServer(http.Dir(staticDir)))).Name("staticHandler")

	d.oauth = NewOAuthTransportStore()

	return nil
}

func (d *Driver) baseURL() (*url.URL, error) {
	return url.Parse(d.ctx.Config().Provider.URL)
}

// creates an error transaction
func (d *Driver) setPayPalError(p *payment.Payment, data []byte) {
	log := d.log.New(log15.Ctx{
		"method":    "setPayPalError",
		"projectID": p.ProjectID(),
		"paymentID": p.ID(),
	})
	log.Warn("status error")

	paypalTx := &Transaction{
		ProjectID: p.ProjectID(),
		PaymentID: p.ID(),
		Timestamp: time.Now(),
		Type:      TransactionTypeError,
	}
	paypalTx.Data = data
	err := InsertTransactionDB(d.ctx.PaymentDB(), paypalTx)
	if err != nil {
		log.Error("error saving paypal transaction", log15.Ctx{"err": err})
	}
}

// execute an HTTP request
func httpDo(
	ctx *service.Context,
	createTr func() (*oauth.Transport, error),
	req *http.Request,
	f func(*http.Response, error) error) error {

	tr, err := createTr()
	if err != nil {
		ctx.Log().Error("error on auth transport", log15.Ctx{"err": err})
		return err
	}
	err = tr.AuthenticateClient()
	if err != nil {
		ctx.Log().Error("error authenticating", log15.Ctx{"err": err})
		return err
	}
	if Debug {
		ctx.Log().Debug("authenticated", log15.Ctx{"accessToken": tr.Token.AccessToken})
	}
	cl := tr.Client()
	c := make(chan error, 1)
	go func() { c <- f(cl.Do(req)) }()
	select {
	case <-ctx.Done():
		if httpTr, ok := tr.Transport.(*http.Transport); ok {
			httpTr.CancelRequest(req)
		}
		<-c
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func (d *Driver) getPayment(p *payment.Payment) {
	log := d.log.New(log15.Ctx{"method": "getPayment"})
	paypalTx, err := TransactionByPaymentIDAndTypeDB(d.ctx.PaymentDB(service.ReadOnly), p.PaymentID(), TransactionTypeCreatePaymentResponse)
	if err != nil {
		log.Error("error retrieving paypal transaction. unitialized payment?", log15.Ctx{"err": err})
		return
	}
	links, err := paypalTx.PayPalLinks()
	if err != nil {
		log.Error("error retrieving paypal links", log15.Ctx{"err": err})
	}
	var selfURL *url.URL
	var req *http.Request
	if selfLink, ok := links["self"]; !ok {
		log.Error("no self link in paypal transaction", log15.Ctx{"links": links})
		return
	} else {
		selfURL, err = url.Parse(selfLink.HRef)
		if err != nil {
			log.Error("error parsing self URL", log15.Ctx{"err": err})
			return
		}
		req, err = http.NewRequest(selfLink.Method, selfURL.String(), nil)
		if err != nil {
			log.Error("error creating HTTP request", log15.Ctx{"err": err})
			return
		}
	}
	method, err := payment_method.PaymentMethodByIDDB(d.ctx.PaymentDB(service.ReadOnly), p.Config.PaymentMethodID.Int64)
	if err != nil {
		log.Error("error retrieving payment method", log15.Ctx{"err": err})
		return
	}
	cfg, err := ConfigByPaymentMethodDB(d.ctx.PaymentDB(service.ReadOnly), method)
	if err != nil {
		log.Error("error retrieving paypal config", log15.Ctx{"err": err})
		return
	}
	responseFunc := func(resp *http.Response, err error) error {
		if err != nil {
			log.Error("error on HTTP call", log15.Ctx{"err": err})
			return err
		}
		respBody, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Error("error reading response body", log15.Ctx{"err": err})
			return err
		}
		log = log.New(log15.Ctx{"responseBody": string(respBody)})
		pay := &PaypalPayment{}
		err = json.Unmarshal(respBody, pay)
		if err != nil {
			log.Error("error decoding response", log15.Ctx{"err": err})
			return err
		}
		paypalTx, err = NewPayPalPaymentTransaction(pay)
		if err != nil {
			log.Error("error creating paypal transaction", log15.Ctx{"err": err})
			return err
		}
		paypalTx.ProjectID = p.ProjectID()
		paypalTx.PaymentID = p.ID()
		paypalTx.Type = TransactionTypeGetPaymentResponse

		err = InsertTransactionDB(d.ctx.PaymentDB(), paypalTx)
		if err != nil {
			log.Error("error saving paypal transaction", log15.Ctx{"err": err})
			return err
		}
		return nil
	}

	err = httpDo(d.ctx, d.oAuthTransportFunc(p, cfg), req, responseFunc)
	if err != nil {
		log.Error("error on executing HTTP request", log15.Ctx{"err": err})
	}
}

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
		return d.statusHandler(currentTx, p, d.InitPageHandler(p)), nil
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

	return d.statusHandler(currentTx, p, d.InitPageHandler(p)), nil
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

		paypalTx, err := NewPayPalPaymentTransaction(paypalP)
		if err != nil && paypalTx == nil {
			log.Error("error on creating response transaction", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrProvider
		}
		if err != nil {
			log.Warn("error on parsing response for transaction", log15.Ctx{"err": err})
		}
		paypalTx.ProjectID = p.ProjectID()
		paypalTx.PaymentID = p.ID()
		paypalTx.Type = TransactionTypeCreatePaymentResponse

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
			log.Crit("error on creating db tx", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrDatabase
		}
		err = InsertTransactionTx(tx, paypalTx)
		if err != nil {
			log.Error("error saving paypal response", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrDatabase
		}
		if paypalP.State != "created" {
			log.Error("invalid paypal state received", log15.Ctx{"state": paypalP.State})
			paypalTx.Type = TransactionTypeError
			paypalTx.Timestamp = time.Now()
			err = InsertTransactionTx(tx, paypalTx)
			if err != nil {
				log.Error("error saving error state", log15.Ctx{"err": err})
				d.setPayPalError(p, respBody)
				return ErrDatabase
			}
		}

		commit = true
		err = tx.Commit()
		if err != nil {
			log.Crit("error on commit", log15.Ctx{"err": err})
			d.setPayPalError(p, respBody)
			return ErrDatabase
		}
		return nil
	}

	err = httpDo(d.ctx, d.oAuthTransportFunc(p, cfg), req, responseFunc)
	if err != nil {
		log.Error("error on create payment request", log15.Ctx{"err": err})
	}
}

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

func (d *Driver) CancelHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "CancelHandler"})
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

		var paymentTx *payment.PaymentTransaction
		var commitIntent paymentService.CommitIntentFunc
		if p.Status != payment.PaymentStatusCancelled {
			if Debug {
				log.Debug("intent cancel")
			}
			paymentTx, commitIntent, err = d.paymentService.IntentCancel(p, 500*time.Millisecond)
			if err != nil {
				log.Error("error on intent payment cancel", log15.Ctx{"err": err})
				d.PaymentErrorHandler(p).ServeHTTP(w, r)
				return
			}
			err = d.paymentService.SetPaymentTransaction(tx, paymentTx)
			if err != nil {
				log.Error("error creating payment transaction", log15.Ctx{"err": err})
				d.InternalErrorHandler(p).ServeHTTP(w, r)
				return
			}
		}

		commit = true
		err = tx.Commit()
		if err != nil {
			log.Crit("error on commit", log15.Ctx{"err": err})
			d.InternalErrorHandler(p).ServeHTTP(w, r)
			return
		}

		// do notify on new payment tx
		if commitIntent != nil {
			if Debug {
				log.Debug("intent commit", log15.Ctx{"commitIntent": commitIntent})
			}
			commitIntent()
		}

		d.CancelPageHandler(p).ServeHTTP(w, r)
	})
}
