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

func (d *Driver) getTemplate(tmplDir, locale, baseName string) (tmpl *template.Template, tmplLocale string, err error) {
	tmplFile, err := getTemplateFile(tmplDir, locale, baseName)
	if err != nil {
		return nil, "", err
	}
	tmplB, err := ioutil.ReadFile(tmplFile)
	if err != nil {
		return nil, "", err
	}
	tmpl = template.New("page")
	tmpl.Funcs(template.FuncMap(map[string]interface{}{
		"staticPath": func() (string, error) {
			url, err := d.mux.Get("staticHandler").URLPath()
			if err != nil {
				return "", err
			}
			return url.Path, nil
		},
	}))
	_, err = tmpl.Parse(string(tmplB))
	if err != nil {
		return nil, "", err
	}
	tmplLocale = path.Base(path.Ext(tmplFile))
	return tmpl, tmplLocale, nil
}

func (d *Driver) InitPageHandler(p *payment.Payment) http.Handler {
	const baseName = "init.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "InitPageHandler"})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmpl, locale, err := d.getTemplate(d.tmplDir, p.Config.Locale.String, baseName)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmplData := make(map[string]interface{})
		tmplData["locale"] = locale
		tmplData["payment"] = p
		tmplData["paymentID"] = d.paymentService.EncodedPaymentID(p.PaymentID())
		tmplData["amount"] = p.DecimalRound(2)
		buf := buffer()
		err = tmpl.Execute(buf, tmplData)
		if err != nil {
			log.Error("error on template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = io.Copy(w, buf)
		putBuffer(buf)
		buf = nil
		if err != nil {
			log.Error("error writing buffered output", log15.Ctx{"err": err})
		}
	})
}

func (d *Driver) InternalErrorHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}
func (d *Driver) PaymentErrorHandler(p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}

func (d *Driver) ReturnHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}
func (d *Driver) CancelHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
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

func (d *Driver) StatusHandler(tx *Transaction, p *payment.Payment) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// will be true when the polling (ajax) should stop and reload
		var h http.Handler
		var cont bool
		switch tx.Type {
		case TransactionTypeError:
			h = d.PaymentErrorHandler(p)
			cont = true
		case TransactionTypeCreatePaymentResponse:
			h = d.ApprovalHandler(tx, p)
			cont = true
		default:
			h = d.InitPageHandler(p)
		}
		// ajax poll?
		if strings.Contains(r.Header.Get("Content-Type"), "application/json") ||
			strings.Contains(r.Header.Get("Accept"), "application/json") {
			w.Header().Set("Content-Type", "text/plain; charset-utf8")
			_, err := fmt.Fprintf(w, "%t", cont)
			if err != nil {
				d.log.Error("error writing response", log15.Ctx{
					"method": "StatusHandler",
					"err":    err,
				})
			}
			return
		}
		h.ServeHTTP(w, r)
	})
}
