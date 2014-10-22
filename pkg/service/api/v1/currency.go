package v1

import (
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/paymentd/currency"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"path"
)

// return a handler brokering the project related admin api requests
func (a *AdminAPI) CurrencyRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log := a.log.New(log15.Ctx{"method": "Currency Request"})
		log.Info("Method:" + r.Method)

		if r.Method == "GET" {

			urlpath, currencyParam := path.Split(r.URL.Path)
			log.Info("urlpath: " + urlpath)
			log.Info("param: " + currencyParam)
			// get all
			if len(currencyParam) == 0 {

				db := a.ctx.PaymentDB(service.ReadOnly)
				cl, err := currency.CurrencyAllDB(db)
				if err != nil {
					log.Error("get all from DB failed.", log15.Ctx{"err": err})
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				je := json.NewEncoder(w)
				err = je.Encode(&cl)
				if err != nil {
					log.Error("json encode failed.", log15.Ctx{"err": err})
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				return
			}

			// get one Currency
			db := a.ctx.PaymentDB(service.ReadOnly)
			if len(currencyParam) != 3 {
				log.Info("param incorrect: " + currencyParam)
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			c, err := currency.CurrencyByCodeISO4217DB(db, currencyParam)
			if err == currency.ErrCurrencyNotFound {
				log.Info("currency " + currencyParam + " not found")
				w.WriteHeader(http.StatusNotFound)
				return
			} else if err != nil {
				log.Info("get currency from DB failed ", log15.Ctx{"err": err})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// output
			je := json.NewEncoder(w)
			err = je.Encode(&c)
			if err != nil {
				log.Error("json encode failed.", log15.Ctx{"err": err})
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

		} else {
			log.Info("not a GET Request: " + r.Method)
			w.WriteHeader(http.StatusBadRequest)
		}

	})
}
