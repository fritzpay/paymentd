package paypal_rest

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"gopkg.in/inconshreveable/log15.v2"
)

var (
	ErrTemplateNotFound    = errors.New("template not found")
	ErrInvalidTemplateFile = errors.New("invalid template file")
)

// defensively try to normalize the locale
//
// it does not necessarily return a valid locale. in this case the default
// locale will be used anyways
func normalizeLocale(l string) string {
	if l == "" {
		return "_"
	}
	l = strings.Replace(l, "-", "_", -1)
	parts := strings.Split(l, "_")
	if len(parts) == 2 {
		parts[0] = strings.ToLower(parts[0])
		parts[1] = strings.ToUpper(parts[1])
		return strings.Join(parts, "_")
	}
	return l
}

func getTemplateFile(tmplDir, locale, baseName string) (string, error) {
	tmplFile := path.Join(tmplDir, normalizeLocale(locale), baseName)
	inf, err := os.Stat(tmplFile)
	if err != nil && os.IsNotExist(err) {
		tmplFile = path.Join(tmplDir, defaultLocale, baseName)
		inf, err = os.Stat(tmplFile)
	}
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrTemplateNotFound
		}
		return "", err
	}
	if inf.IsDir() {
		return "", ErrInvalidTemplateFile
	}
	return tmplFile, nil
}

func (d *Driver) getTemplate(tmpl *template.Template, tmplDir, locale, baseName string) (err error) {
	tmplFile, err := getTemplateFile(tmplDir, locale, baseName)
	if err != nil {
		return err
	}
	tmplB, err := ioutil.ReadFile(tmplFile)
	if err != nil {
		return err
	}
	tmplLocale := path.Base(path.Ext(tmplFile))
	tmpl.Funcs(template.FuncMap(map[string]interface{}{
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
	_, err = tmpl.Parse(string(tmplB))
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

func writeTemplateBuf(log log15.Logger, w io.Writer, tmpl *template.Template, tmplData interface{}) error {
	buf := buffer()
	err := tmpl.Execute(buf, tmplData)
	if err != nil {
		log.Error("error on template", log15.Ctx{"err": err})
		return ErrInternal
	}
	_, err = io.Copy(w, buf)
	putBuffer(buf)
	buf = nil
	if err != nil {
		log.Error("error writing buffered output", log15.Ctx{"err": err})
	}
	return nil
}

// InitPageHandler serves the init page (loading screen)
func (d *Driver) InitPageHandler(p *payment.Payment) http.Handler {
	const baseName = "init.html.tmpl"
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
		err = writeTemplateBuf(log, w, tmpl, tmplData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

// InternalErrorHandler serves the page notifying the user about a (critical)
// internal error. The payment can not continue.
//
// It can handle a nil payment parameter.
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
		writeTemplateBuf(log, w, tmpl, tmplData)
	})
}

func (d *Driver) PaymentErrorHandler(p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
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
		writeTemplateBuf(log, w, tmpl, tmplData)
	})
}

func (d *Driver) CancelPageHandler(p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{
			"method":    "CancelPageHandler",
			"projectID": p.ProjectID(),
			"paymentID": p.PaymentID(),
		})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.New("cancel")
		const baseName = "cancel.html.tmpl"
		err := d.getTemplate(tmpl, d.tmplDir, p.Config.Locale.String, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmplData := d.templatePaymentData(p)
		err = writeTemplateBuf(log, w, tmpl, tmplData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
}

func (d *Driver) ReturnPageHandler(p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{
			"method":    "ReturnPageHandler",
			"projectID": p.ProjectID(),
			"paymentID": p.PaymentID(),
		})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl := template.New("return")
		const baseName = "return.html.tmpl"
		err := d.getTemplate(tmpl, d.tmplDir, p.Config.Locale.String, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmplData := d.templatePaymentData(p)
		err = writeTemplateBuf(log, w, tmpl, tmplData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
		writeTemplateBuf(log, w, tmpl, tmplData)
	})
}

func (d *Driver) ApprovalHandler(tx *Transaction, p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{
			"method":               "ApprovalHandler",
			"projectID":            p.ProjectID(),
			"paymentID":            p.PaymentID(),
			"transactionTimestamp": tx.Timestamp.UnixNano(),
		})
		links, err := tx.PayPalLinks()
		if err != nil {
			log.Error("transaction links error", log15.Ctx{"err": err})
			d.PaymentErrorHandler(p).ServeHTTP(w, r)
			return
		}
		if links["approval_url"] == nil {
			log.Error("no approval URL")
			d.PaymentErrorHandler(p).ServeHTTP(w, r)
			return
		}
		http.Redirect(w, r, links["approval_url"].HRef, http.StatusTemporaryRedirect)
	})
}

func (d *Driver) PaymentStatusHandler(p *payment.Payment) http.Handler {
	switch p.Status {
	case payment.PaymentStatusCancelled:
		return d.CancelPageHandler(p)
	case payment.PaymentStatusPaid:
		return d.SuccessHandler(p)
	case payment.PaymentStatusError:
		return d.PaymentErrorHandler(p)
	default:
		d.log.Warn("unknown payment status", log15.Ctx{
			"method":                   "PaymentStatusHandler",
			"paymentTransactionStatus": p.Status,
		})
		return d.PaymentErrorHandler(p)
	}
}

// the returned handler will serve the appropriate init action based on the current
// paypal transaction status
func (d *Driver) statusHandler(tx *Transaction, p *payment.Payment, defaultHandler http.Handler) http.Handler {
	return d.pollStatusHandler(tx, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch tx.Type {
		case TransactionTypeCreatePaymentResponse:
			if tx.PaypalState.String == "created" {
				d.ApprovalHandler(tx, p).ServeHTTP(w, r)
				return
			}
			d.PaymentStatusHandler(p).ServeHTTP(w, r)
		case TransactionTypeError, TransactionTypeCancelled, TransactionTypeExecutePaymentResponse:
			d.PaymentStatusHandler(p).ServeHTTP(w, r)
		default:
			defaultHandler.ServeHTTP(w, r)
		}
	}))
}

func (d *Driver) pollStatusHandler(tx *Transaction, parent http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ajax poll?
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") &&
			!strings.Contains(r.Header.Get("Accept"), "application/json") {
			parent.ServeHTTP(w, r)
			return
		}
		// will be true when the polling (ajax) should stop and reload
		cont := true
		switch tx.Type {
		// wait on create payment request
		case TransactionTypeCreatePayment:
			cont = false
			// wait on execute payment request
		case TransactionTypeExecutePayment:
			cont = false
		}
		w.Header().Set("Content-Type", "application/json")
		_, err := fmt.Fprintf(w, "{\"c\": %t}", cont)
		if err != nil {
			d.log.Error("error writing response", log15.Ctx{
				"method": "StatusHandler",
				"err":    err,
			})
		}
		return
	})
}
