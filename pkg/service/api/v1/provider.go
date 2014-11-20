package v1

import (
	"net/http"

	"github.com/fritzpay/paymentd/pkg/paymentd/provider"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

type ProviderAdminAPIResponse struct {
	AdminAPIResponse
}

// return a handler brokering get provider by give id
func (a *AdminAPI) ProviderGetRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{"method": "Provider Request"})

		if r.Method != "GET" {
			ErrInval.Write(w)
			log.Info("unsupported method " + r.Method)
		}

		vars := mux.Vars(r)
		providerParam := vars["provider"]

		// get one Provider
		db := a.ctx.PaymentDB(service.ReadOnly)
		pr, err := provider.ProviderByNameDB(db, providerParam)
		if err == provider.ErrProviderNotFound {
			ErrNotFound.Write(w)
			log.Info("provider not found", log15.Ctx{"providerName": providerParam})
			return
		} else if err != nil {
			ErrDatabase.Write(w)
			log.Error("database error", log15.Ctx{"err": err})
			return
		}

		resp := ProviderAdminAPIResponse{}
		resp.Info = "provider " + pr.Name + " found."
		resp.Status = StatusSuccess
		resp.Response = pr
		// response write
		resp.Write(w)
		if err != nil {
			log.Error("write error", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	})
}

// return a handler brokering get a provider
func (a *AdminAPI) ProviderGetAllRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		// get all
		log := a.log.New(log15.Ctx{"method": "Provider Request"})
		db := a.ctx.PaymentDB(service.ReadOnly)
		prl, err := provider.ProviderAllDB(db)
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("database error", log15.Ctx{"err": err})
			return
		}
		// response write
		resp := ProviderAdminAPIResponse{}
		resp.Status = StatusSuccess
		resp.Info = "providers found"
		resp.Response = prl
		resp.Write(w)
		if err != nil {
			log.Error("write error", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		return
	})
}
