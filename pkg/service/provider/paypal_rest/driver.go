package paypal_rest

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

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
