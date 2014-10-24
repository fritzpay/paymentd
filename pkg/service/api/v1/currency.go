package v1

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/currency"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"path"
)

type CurrencyAdminAPIResponse struct {
	AdminAPIResponse
}

// return a handler brokering get all currencies
func (a *AdminAPI) CurrencyGetRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{"method": "Currency Request"})

		if r.Method != "GET" {
			ErrInval.Write(w)
			log.Info("unsupported method " + r.Method)
		}

		urlpath, currencyParam := path.Split(r.URL.Path)
		log.Info("urlpath: " + urlpath)
		log.Info("param: " + currencyParam)

		// get one Currency
		if len(currencyParam) != 3 {
			ErrReadParam.Write(w)
			log.Info("malformed param: " + currencyParam)
			return
		}

		db := a.ctx.PaymentDB(service.ReadOnly)
		c, err := currency.CurrencyByCodeISO4217DB(db, currencyParam)
		if err == currency.ErrCurrencyNotFound {
			ErrNotFound.Write(w)
			log.Info("currency " + currencyParam + " not found")
			return
		} else if err != nil {
			ErrDatabase.Write(w)
			log.Error("database error ", log15.Ctx{"err": err})
			return
		}

		resp := CurrencyAdminAPIResponse{}
		resp.Info = "currency " + c.CodeISO4217 + " found"
		resp.Status = StatusSuccess
		resp.Response = c
		// response write
		resp.Write(w)
		if err != nil {
			log.Error("write error", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	})
}

// return a handler brokering get a currency
func (a *AdminAPI) CurrencyGetAllRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// get all
		log := a.log.New(log15.Ctx{"method": "Currency Request"})
		db := a.ctx.PaymentDB(service.ReadOnly)
		cl, err := currency.CurrencyAllDB(db)
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("database error", log15.Ctx{"err": err})
			return
		}
		// response write
		resp := CurrencyAdminAPIResponse{}
		resp.Status = StatusSuccess
		resp.Info = "currencies found"
		resp.Response = cl
		resp.Write(w)
		if err != nil {
			log.Error("write error", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	})
}
