package admin

import (
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

// API represents the admin API in version 1.x
type API struct {
	router *mux.Router

	log log15.Logger
}

// NewAPI creates a new admin API
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
