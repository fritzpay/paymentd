package api

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/api/v1"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

// Handler is the (HTTP) API Handler
type Handler struct {
	ctx *service.Context
	log log15.Logger

	mux *mux.Router
}

// NewHandler creates a new API Handler
func NewHandler(ctx *service.Context) (*Handler, error) {
	h := &Handler{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api",
		}),

		mux: mux.NewRouter(),
	}

	h.log.Info("registering API service v1...")
	v1.NewService(h.ctx, h.mux)
	v1.Log = h.log.New(log15.Ctx{
		"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1",
	})

	return h, nil
}

// ServeHTTP implements the http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			h.log.Crit("panic on serving HTTP", log15.Ctx{"panic": err})
		}
	}()
	service.SetRequestContext(r, h.ctx)
	defer service.ClearRequestContext(r)
	h.mux.ServeHTTP(w, r)
}
