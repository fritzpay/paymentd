package fritzpay

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"os"
	"path"
	"time"
)

const (
	FritzpayDriverPath = "/fritzpay"
)

const (
	defaultLocale = "en_US"
)

var (
	ErrDB       = errors.New("database error")
	ErrConflict = errors.New("conflict")
)

type Driver struct {
	ctx     *service.Context
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
	d.tmplDir = path.Join(cfg.Provider.ProviderTemplateDir, "fritzpay")
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

	mux.HandleFunc(FritzpayDriverPath+"/status", d.Status)
	mux.Handle(FritzpayDriverPath+"/payment", d.PaymentInfo())
	return nil
}

func (d *Driver) InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error) {
	log := d.log.New(log15.Ctx{
		"method":          "InitPayment",
		"projectID":       p.ProjectID(),
		"paymentID":       p.ID(),
		"paymentMethodID": method.ID,
	})
	if Debug {
		log.Debug("initialize payment")
	}
	if method.Status != payment_method.PaymentMethodStatusActive {
		log.Warn("payment requested with inactive payment method")
		return nil, fmt.Errorf("inactive payment method id %d", method.ID)
	}

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
	tx, err = d.ctx.PaymentDB().Begin()
	if err != nil {
		commit = true
		log.Crit("error on begin tx", log15.Ctx{"err": err})
		return nil, ErrDB
	}
	fritzpayP, err := PaymentByPaymentIDTx(tx, p.PaymentID())
	if err != nil && err != ErrPaymentNotFound {
		log.Error("error retrieving payment id", log15.Ctx{"err": err})
		return nil, ErrDB
	}
	// payment does already exist
	if err == nil {
		if fritzpayP.MethodKey != method.MethodKey {
			log.Crit("payment does exist but has a different method key", log15.Ctx{
				"registeredMethodKey": fritzpayP.MethodKey,
				"requestMethodKey":    method.MethodKey,
			})
			return nil, ErrConflict
		}
	}
	if err == ErrPaymentNotFound {
		// create new fritzpay payment
		fritzpayP.ProjectID = p.ProjectID()
		fritzpayP.PaymentID = p.ID()
		fritzpayP.Created = time.Now()
		fritzpayP.MethodKey = method.MethodKey
		err = InsertPaymentTx(tx, &fritzpayP)
		if err != nil {
			log.Error("error creating new payment", log15.Ctx{"err": err})
			return nil, ErrDB
		}
	}
	log = log.New(log15.Ctx{"fritzpayPaymentID": fritzpayP.ID})

	if currentStatus, err := d.paymentService.PaymentTransaction(tx, p); err != nil && err != payment.ErrPaymentTransactionNotFound {
		log.Error("error retrieving payment transaction", log15.Ctx{"err": err})
		return nil, ErrDB
	} else {
		if currentStatus.Status != payment.PaymentStatusPending {
			paymentTx := p.NewTransaction(payment.PaymentStatusPending)
			paymentTx.Amount = 0
			paymentTx.Comment.String, paymentTx.Comment.Valid = "initialized by FritzPay demo provider", true
			err = d.paymentService.SetPaymentTransaction(tx, paymentTx)
			if err != nil {
				log.Error("error setting payment tx", log15.Ctx{"err": err})
				return nil, ErrDB
			}
		}
	}
	err = tx.Commit()
	commit = true
	if err != nil {
		log.Crit("error on commit", log15.Ctx{"err": err})
		return nil, ErrDB
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	}), nil
}

func (d *Driver) Status(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "FritzPay OK.")
}

func (d *Driver) PaymentInfo() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	})
}
