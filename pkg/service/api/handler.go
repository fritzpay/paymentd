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

	httpHandler http.Handler
}

// NewHandler creates a new API Handler
func NewHandler(ctx *service.Context) (*Handler, error) {
	h := &Handler{
		ctx: ctx,

		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api",
		}),
	}

	mux := http.NewServeMux()

	h.log.Info("registering API service v1...")
	mux.Handle(v1.ServicePath, v1.NewService(h.ctx))

	return h, nil
}

// ServeHTTP implements the http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.httpHandler.ServeHTTP(w, r)
}
