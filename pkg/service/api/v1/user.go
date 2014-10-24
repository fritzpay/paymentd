package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"net/http"
)

type UserAdminAPIResponse struct {
	AdminAPIResponse
}

// GetUserID returns a utility handler. This endpoint displays the user ID, which is stored
// in the authorization container
func (a *AdminAPI) GetUserID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "GET" {
			ErrMethod.Write(w)
			return
		}
		ctx := service.RequestContext(r)
		auth := ctx.Value(service.ContextVarAuthKey).(map[string]interface{})

		resp := UserAdminAPIResponse{}
		resp.Info = "user id"
		resp.Status = StatusSuccess
		resp.Response = auth[AuthUserIDKey].(string)
		resp.Write(w)

	})
}
