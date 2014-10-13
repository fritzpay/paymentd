package admin

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"strings"
)

// API represents the admin API in version 1.x
type API struct {
	ctx *service.Context
	log log15.Logger
}

// NewAPI creates a new admin API
func NewAPI(ctx *service.Context) *API {
	a := &API{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1/admin"}),
	}
	return a
}

// Handler wraps the given handler with admin API related actions
func (a *API) Handler(parent http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.Split(r.URL.Path, "/")
		if path[len(path)-1] == "" {
			path = path[:len(path)-1]
		}
		if path[len(path)-1] == "test" {
			w.Write([]byte("test"))
			return
		}
		parent.ServeHTTP(w, r)
	})
}
