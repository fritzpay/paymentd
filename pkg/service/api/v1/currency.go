package v1

import (
	"net/http"

	"github.com/fritzpay/paymentd/pkg/paymentd/currency"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

type CurrencyAdminAPIResponse struct {
	AdminAPIResponse
}

// return a handler brokering get all currencies
func (a *AdminAPI) CurrencyGetRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{"method": "CurrencyGetRequest"})

		// get param
		vars := mux.Vars(r)
		currencyParam := vars["currencycode"]
		if r.Method != "GET" {
			ErrInval.Write(w)
			log.Info("unsupported method", log15.Ctx{"requestMethod": r.Method})
			return
		}

		// get one Currency
		if len(currencyParam) != 3 {
			ErrReadParam.Write(w)
			log.Info("malformed param", log15.Ctx{"currencyParam": currencyParam})
			return
		}

		log = log.New(log15.Ctx{"currencyParam": currencyParam})

		db := a.ctx.PaymentDB(service.ReadOnly)
		c, err := currency.CurrencyByCodeISO4217DB(db, currencyParam)
		if err == currency.ErrCurrencyNotFound {
			ErrNotFound.Write(w)
			log.Info("currency not found")
			return
		} else if err != nil {
			ErrDatabase.Write(w)
			log.Error("database error", log15.Ctx{"err": err})
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
		log := a.log.New(log15.Ctx{"method": "CurrencyGetAllRequest"})

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
