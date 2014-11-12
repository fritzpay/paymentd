package paypal_rest

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	PaypalDriverPath = "/paypal"
)

const (
	providerTemplateDir = "paypal_rest"
)

var (
	ErrDatabase = errors.New("database error")
	ErrInternal = errors.New("paypal driver internal error")
)

type Driver struct {
	ctx     *service.Context
	mux     *mux.Router
	log     log15.Logger
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

	d.mux = mux

	d.oauth = NewOAuthTransportStore()

	return nil
}
