package admin

import (
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

type API struct {
	router *mux.Router

	log log15.Logger
}

func NewAPI(r *mux.Router, log log15.Logger) *API {
	a := &API{
		router: r,

		log: log.New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1/admin"}),
	}

	a.router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test here"))
	})
	return a
}
