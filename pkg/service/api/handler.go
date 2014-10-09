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
	router *mux.Router
	ctx    *service.Context

	log log15.Logger
}

// NewHandler creates a new API Handler
func NewHandler(ctx *service.Context) (*Handler, error) {
	h := &Handler{
		router: mux.NewRouter(),
		ctx:    ctx,

		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api",
		}),
	}

	h.log.Info("registering API service v1...")
	v1.NewService(h.ctx, h.router)

	return h, nil
}

// ServeHTTP implements the http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
