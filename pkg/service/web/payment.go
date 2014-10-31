package web

import (
	"net/http"
)

func (h *Handler) PaymentHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	})
}
