package stripe

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	tmpl "github.com/fritzpay/paymentd/pkg/template"
	"github.com/gorilla/mux"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	// StripeDriverPath is the (sub-)path under which Stripe driver endpoints
	// will be attached
	StripeDriverPath = "/stripe"
)

const (
	providerTemplateDir  = "stripe"
	defaultLocale        = "en_US"
	stripeSecretKey      = ""
	stripePublishableKey = ""
)

var (
	ErrDatabase = errors.New("database error")
	ErrInternal = errors.New("stripe driver internal error")
	ErrHTTP     = errors.New("HTTP error")
	ErrProvider = errors.New("provider error")
)

// Driver is the Stripe provider driver
type Driver struct {
	context        *service.Context
	tmplDir        string
	log            log15.Logger
	mux            *mux.Router
	paymentService *paymentService.Service
}

func (d *Driver) Attach(ctx *service.Context, m *mux.Router) error {

	d.context = ctx
	d.log = ctx.Log().New(log15.Ctx{
		"pkg": "github.com/fritzpay/paymentd/pkg/service/provider/stripe",
	})

	//set template path
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

	d.paymentService, err = paymentService.NewService(ctx)

	// add subrouting
	driverRoute := m.PathPrefix(StripeDriverPath)
	url, err := driverRoute.URLPath()
	if err != nil {
		d.log.Error("error determining path prefix", log15.Ctx{"err": err})
		return fmt.Errorf("error on subroute path: %v", err)
	}
	d.mux = driverRoute.Subrouter()
	d.mux.Handle("/process", ctx.RateLimitHandler(d.ProcessHandler())).Name("processFormHandler")
	staticDir := path.Join(d.tmplDir, "static")
	d.log.Info("serving static dir", log15.Ctx{
		"staticDir": staticDir,
		"prefix":    url.Path + "/static",
	})
	d.mux.PathPrefix("/static").Handler(http.StripPrefix(url.Path+"/static", http.FileServer(http.Dir(staticDir)))).Name("staticHandler")

	if err != nil {
		d.log.Error("error initializing payment service", log15.Ctx{"err": err})
		return err
	}

	return err
}

func (d *Driver) InitPayment(p *payment.Payment, pm *payment_method.Method) (http.Handler, error) {

	// start transaction
	// show stripe.js form
	return d.InitPageHandler(p), nil

}

// InitPageHandler serves the init page (loading screen)
func (d *Driver) InitPageHandler(p *payment.Payment) http.Handler {
	const baseName = "form.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "InitPageHandler"})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.New("init")
		err := d.getTemplate(tmpl, d.tmplDir, p.Config.Locale.String, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmplData := d.templatePaymentData(p)
		err = tmpl.Execute(w, tmplData)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

// takes the post request and handles the stripe checkout
func (d *Driver) ProcessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "ProcessHandler"})

		r.ParseForm()
		paymentIDStr := r.Form.Get("paymentid")
		stripeTokenStr := r.Form.Get("stripeToken")

		paymentID, err := payment.ParsePaymentIDStr(paymentIDStr)
		if err != nil {
			log.Warn("error parsing payment ID", log15.Ctx{
				"err":          err,
				"paymentIDStr": paymentIDStr,
			})
			d.BadRequestHandler().ServeHTTP(w, r)
			return
		}

		// get payment data
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
		tx, err = d.context.PaymentDB().Begin()
		paymentID = d.paymentService.DecodedPaymentID(paymentID)

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

		// stripe charge
		stripe.Key = stripeSecretKey

		params := &stripe.ChargeParams{
			Amount:   uint64(p.Amount),
			Currency: stripe.Currency(p.Currency),
			Card: &stripe.CardParams{
				Token: stripeTokenStr,
			},
		}
		ch, err := charge.New(params)
		if err != nil {
			log.Error("error retrieving stripe charge object", log15.Ctx{"err": err})
			d.InternalErrorHandler(nil).ServeHTTP(w, r)
		}
		log.Debug("payment", log15.Ctx{"payment": p})
		log.Debug("charge params", log15.Ctx{"params": params})
		log.Debug("charge object", log15.Ctx{"charge": ch})
		// check charge

		// log stripe charge
		d.SuccessHandler(p).ServeHTTP(w, r)
		commit = true
		err = tx.Commit()
		if err != nil {
			log.Crit("error on commit", log15.Ctx{"err": err})
			d.InternalErrorHandler(p)

		}
	})
}

// ProcessFormPageHandler serves the post action (form processing)
func (d *Driver) processFormPageHandler(p *payment.Payment) http.Handler {
	const baseName = "form.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "InitPageHandler"})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.New("init")
		err := d.getTemplate(tmpl, d.tmplDir, p.Config.Locale.String, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmplData := d.templatePaymentData(p)
		err = tmpl.Execute(w, tmplData)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (d *Driver) getTemplate(t *template.Template, tmplDir, locale, baseName string) (err error) {
	tmplFile, err := tmpl.TemplateFileName(tmplDir, locale, defaultLocale, baseName)
	if err != nil {
		return err
	}
	tmplB, err := ioutil.ReadFile(tmplFile)
	if err != nil {
		return err
	}
	tmplLocale := path.Base(path.Ext(tmplFile))
	t.Funcs(template.FuncMap(map[string]interface{}{
		"staticPath": func() (string, error) {
			url, err := d.mux.Get("staticHandler").URLPath()
			if err != nil {
				return "", err
			}
			return url.Path, nil
		},
		"locale": func() string {
			return tmplLocale
		},
	}))
	_, err = t.Parse(string(tmplB))
	if err != nil {
		return err
	}
	return nil
}

func (d *Driver) templatePaymentData(p *payment.Payment) map[string]interface{} {
	tmplData := make(map[string]interface{})
	if p != nil {
		tmplData["payment"] = p
		tmplData["paymentID"] = d.paymentService.EncodedPaymentID(p.PaymentID())
		tmplData["amount"] = p.DecimalRound(2)

	}
	tmplData["timestamp"] = time.Now().Unix()
	return tmplData

}

func (d *Driver) BadRequestHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func (d *Driver) NotFoundHandler(p *payment.Payment) http.Handler {
	const baseName = "not_found.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "NotFoundHandler"})

		tmplData := d.templatePaymentData(p)
		// do log so we can find the timestamp in the logs
		log.Warn("payment not found", log15.Ctx{"timestamp": tmplData["timestamp"]})
		w.WriteHeader(http.StatusNotFound)
		locale := defaultLocale
		if p != nil {
			locale = p.Config.Locale.String
		}
		tmpl := template.New("not_found")
		err := d.getTemplate(tmpl, d.tmplDir, locale, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			return
		}
		err = tmpl.Execute(w, tmplData)
		if err != nil {
			log.Error("error executing template", log15.Ctx{"err": err})
		}
	})
}

func (d *Driver) InternalErrorHandler(p *payment.Payment) http.Handler {
	const baseName = "internal_error.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "InternalErrorHandler"})

		tmplData := d.templatePaymentData(p)
		// do log so we can find the timestamp in the logs
		log.Error("internal error", log15.Ctx{"timestamp": tmplData["timestamp"]})
		w.WriteHeader(http.StatusInternalServerError)
		locale := defaultLocale
		if p != nil {
			locale = p.Config.Locale.String
		}
		tmpl := template.New("internal_error")
		err := d.getTemplate(tmpl, d.tmplDir, locale, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			return
		}
		err = tmpl.Execute(w, tmplData)
		if err != nil {
			log.Error("error executing template", log15.Ctx{"err": err})
		}
	})
}

func (d *Driver) SuccessHandler(p *payment.Payment) http.Handler {
	const baseName = "success.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "SuccessHandler"})

		tmplData := d.templatePaymentData(p)
		locale := defaultLocale
		if p != nil {
			locale = p.Config.Locale.String
		}
		tmpl := template.New("success")
		err := d.getTemplate(tmpl, d.tmplDir, locale, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			return
		}
		err = tmpl.Execute(w, tmplData)
		if err != nil {
			log.Error("error executing template", log15.Ctx{"err": err})
		}
	})
}
