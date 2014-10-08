package api

import (
	"github.com/fritzpay/paymentd/pkg/config"
	"github.com/fritzpay/paymentd/pkg/service/api/v1"
	"github.com/gorilla/mux"
	"net/http"
)

// Handler is the (HTTP) API Handler
type Handler struct {
	router *mux.Router
}

// NewHandler creates a new API Handler
func NewHandler(cfg config.Config) (*Handler, error) {
	h := &Handler{
		router: mux.NewRouter(),
	}

	v1.NewService(cfg, h.router)

	return h, nil
}

// ServeHTTP implements the http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
