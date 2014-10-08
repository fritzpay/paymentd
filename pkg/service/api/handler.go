package api

import (
	"net/http"
)

// Handler is the API (HTTP) Handler
type Handler struct {
}

// ServeHTTP implements the http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
