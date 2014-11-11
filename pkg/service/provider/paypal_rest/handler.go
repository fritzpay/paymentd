package paypal_rest

import (
	"net/http"
)

func (d *Driver) InternalErrorHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
}
