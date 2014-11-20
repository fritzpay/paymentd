package fritzpay

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	FritzpayDriverPath  = "/fritzpay"
	FritzpayTemplateDir = "fritzpay"
)

const (
	providerIDFritzpay     = "fritzpay"
	defaultLocale          = "en_US"
	fritzpayDefaultTimeout = 30 * time.Second
)

var (
	ErrDB       = errors.New("database error")
	ErrConflict = errors.New("conflict")
)

type Driver struct {
	ctx     *service.Context
	mux     *mux.Router
	log     log15.Logger
	tmplDir string

	paymentService *paymentService.Service
}

func (d *Driver) Attach(ctx *service.Context, mux *mux.Router) error {
	d.ctx = ctx
	d.log = ctx.Log().New(log15.Ctx{
		"pkg": "github.com/fritzpay/paymentd/pkg/service/provider/fritzpay",
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
	d.tmplDir = path.Join(cfg.Provider.ProviderTemplateDir, FritzpayTemplateDir)
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

	d.mux = mux
	mux.HandleFunc(FritzpayDriverPath+"/status", d.Status)
	mux.Handle(FritzpayDriverPath+"/payment", d.PaymentInfo())
	mux.HandleFunc(FritzpayDriverPath+"/f", d.Callback).Name("fritzpayCallback")
	return nil
}

func (d *Driver) Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "FritzPay OK.")
}

func (d *Driver) PaymentInfo() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}

// Callback handles callback from the "psp" (payment service provider; in this case
// a mock implementation)
//
// It will always answer with a HTTP status 200 OK unless there was a data error
// We expect the PSP to re-send the callback notification if we answer with anything
// other than 200
func (d *Driver) Callback(w http.ResponseWriter, r *http.Request) {
	log := d.log.New(log15.Ctx{
		"method": "Callback",
	})
	if Debug {
		log.Debug("received callback", log15.Ctx{"query": r.URL.Query()})
	}
	// always answer with ok
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	paymentIDStr := r.URL.Query().Get("paymentID")
	if paymentIDStr == "" {
		log.Warn("no payment id in callback")
		w.WriteHeader(http.StatusOK)
		return
	}
	paymentID, err := payment.ParsePaymentIDStr(paymentIDStr)
	if err != nil {
		log.Warn("invalid payment id", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusOK)
		return
	}
	log = log.New(log15.Ctx{
		"displayPaymentId": paymentID.String(),
	})
	paymentID = d.paymentService.DecodedPaymentID(paymentID)
	p, err := payment.PaymentByIDDB(d.ctx.PaymentDB(service.ReadOnly), paymentID)
	if err != nil {
		if err == payment.ErrPaymentNotFound {
			log.Warn("payment not found")
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Error("error retrieving payment", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log = log.New(log15.Ctx{
		"projectID": p.ProjectID(),
		"paymentID": p.ID(),
	})
	if !p.Config.IsConfigured() {
		log.Warn("received callback for unconfigured payment")
		w.WriteHeader(http.StatusOK)
		return
	}
	if !p.Config.PaymentMethodID.Valid {
		log.Warn("received callback for payment without payment method")
		w.WriteHeader(http.StatusOK)
		return
	}
	method, err := payment_method.PaymentMethodByIDDB(d.ctx.PaymentDB(service.ReadOnly), p.Config.PaymentMethodID.Int64)
	if err != nil {
		if err == payment_method.ErrPaymentMethodNotFound {
			log.Warn("payment method not found", log15.Ctx{"paymentMethodID": p.Config.PaymentMethodID.Int64})
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Error("error retrieving payment method", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if !method.Active() {
		log.Warn("inactive payment method")
		w.WriteHeader(http.StatusOK)
		return
	}
	if method.Provider.Name != providerIDFritzpay {
		log.Warn("invalid payment method provider", log15.Ctx{"providerName": method.Provider.Name})
		w.WriteHeader(http.StatusOK)
		return
	}
	log = log.New(log15.Ctx{
		"paymentMethodID": method.ID,
		"methodKey":       method.MethodKey,
	})

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
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fritzpayP, err := PaymentByPaymentIDTx(tx, p.PaymentID())
	if err != nil {
		if err == ErrPaymentNotFound {
			log.Warn("callback on unknown payment")
			w.WriteHeader(http.StatusOK)
			return
		}
		log.Error("error retrieving payment", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	currentTx, err := PaymentTransactionCurrentByPaymentIDTx(tx, fritzpayP.ID)
	if err != nil && err != ErrTransactionNotFound {
		log.Error("error retrieving payment tx", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fritzpayTx := PaymentTransaction{
		FritzpayPaymentID: fritzpayP.ID,
		Timestamp:         time.Now(),
	}
	if r.URL.Query().Get("fritzpayID") != "" {
		fritzpayTx.FritzpayID.String, fritzpayTx.FritzpayID.Valid = r.URL.Query().Get("fritzpayID"), true
	}
	switch r.URL.Query().Get("status") {
	case TransactionPSPInit:
		if currentTx.Status == TransactionOpen {
			// noop
			w.WriteHeader(http.StatusOK)
			return
		}
		fritzpayTx.Status = TransactionOpen
	default:
		log.Warn("invalid status", log15.Ctx{"status": r.URL.Query().Get("status")})
		w.WriteHeader(http.StatusOK)
		return
	}
	err = InsertPaymentTransactionTx(tx, fritzpayTx)
	if err != nil {
		log.Error("error on insert payment tx", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	commit = true
	err = tx.Commit()
	if err != nil {
		log.Crit("error on commit", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
