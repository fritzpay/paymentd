package api

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/api/v1"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

// Handler is the (HTTP) API Handler
type Handler struct {
	ctx *service.Context
	log log15.Logger

	mux *http.ServeMux

	requestContexts map[*http.Request]*service.Context
}

// NewHandler creates a new API Handler
func NewHandler(ctx *service.Context) (*Handler, error) {
	h := &Handler{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api",
		}),

		mux: http.NewServeMux(),

		requestContexts: make(map[*http.Request]*service.Context),
	}

	h.log.Info("registering API service v1...")
	h.mux.Handle(v1.ServicePath, v1.NewService(h.ctx).Handler())

	return h, nil
}

// ServeHTTP implements the http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			h.log.Crit("panic on serving HTTP", log15.Ctx{"panic": err})
		}
	}()
	h.mux.ServeHTTP(w, r)
}
