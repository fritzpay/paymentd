package api

import (
	"code.google.com/p/go.net/context"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/service/api/v1"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

// Handler is the (HTTP) API Handler
type Handler struct {
	router *mux.Router
	ctx    context.Context

	log log15.Logger
}

// NewHandler creates a new API Handler
func NewHandler(ctx context.Context) (*Handler, error) {
	var log log15.Logger
	var ok bool
	if log, ok = ctx.Value("log").(log15.Logger); !ok {
		return nil, fmt.Errorf("invalid context. require logger, got %T", ctx.Value("log"))
	}
	h := &Handler{
		router: mux.NewRouter(),

		log: log.New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api",
		}),
	}

	h.ctx = context.WithValue(ctx, "router", h.router)

	h.log.Info("registering API service v1...")
	v1.NewService(h.ctx)

	return h, nil
}

// ServeHTTP implements the http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
