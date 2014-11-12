package paypal_rest

import (
	"errors"
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

func normalizeLocale(l string) string {
	return strings.Replace(l, "-", "_", -1)
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

func getTemplate(tmplFile string) (*template.Template, error) {
	tmplB, err := ioutil.ReadFile(tmplFile)
	if err != nil {
		return nil, err
	}
	tmpl := template.New("page")
	_, err = tmpl.Parse(string(tmplB))
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func (d *Driver) InitPageHandler(p *payment.Payment) http.Handler {
	const baseName = "init.html.tmpl"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := d.log.New(log15.Ctx{"method": "InitPageHandler"})
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		tmplFile, err := getTemplateFile(d.tmplDir, p.Config.Locale.String, baseName)
		if err != nil {
			log.Error("error retrieving template file name", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmpl, err := getTemplate(tmplFile)
		if err != nil {
			log.Error("error initializing template", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tmplData := make(map[string]interface{})
		tmplData["payment"] = p
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
