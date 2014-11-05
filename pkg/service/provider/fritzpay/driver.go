package fritzpay

import (
	"code.google.com/p/go.net/context"
	"database/sql"
	"errors"
	"fmt"
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

func (d *Driver) Callback(w http.ResponseWriter, r *http.Request) {

}

func doInit(ctx context.Context, fritzpayP Payment, callbackURL string) {
	if deadline, ok := ctx.Deadline(); ok {
		// let's assume we will need at least 3 seconds to run
		if deadline.Before(time.Now().Add(3 * time.Second)) {
			return
		}
	}
	log := ctx.Value("log").(log15.Logger).New(log15.Ctx{
		"pkg":         "github.com/fritzpay/paymentd/pkg/service/provider/fritzpay",
		"method":      "doInit",
		"callbackURL": callbackURL,
	})
	tx, err := ctx.Value("paymentDB").(*sql.DB).Begin()
	if err != nil {
		log.Crit("error on begin tx", log15.Ctx{"err": err})
		return
	}
	if Debug {
		log.Debug("worker start...")
	}
	ok := make(chan struct{})
	select {
	case <-ctx.Done():
		log.Warn("cancelled worker", log15.Ctx{"err": ctx.Err()})
		err = tx.Rollback()
		if err != nil {
			log.Crit("error on rollback", log15.Ctx{"err": err})
		}
		return
	case <-ok:
		if Debug {
			log.Debug("worker done")
		}
		return
	}
}
