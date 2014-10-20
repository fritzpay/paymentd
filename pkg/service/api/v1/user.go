package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"net/http"
)

// GetUserID returns a utility handler. This endpoint displays the user ID, which is stored
// in the authorization container
func (a *AdminAPI) GetUserID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ctx := service.RequestContext(r)
		auth := ctx.Value(service.ContextVarAuthKey).(map[string]interface{})
		w.Write([]byte(auth[AuthUserIDKey].(string)))
	})
}
