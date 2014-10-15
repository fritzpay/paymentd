package admin

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"time"
)

const (
	badAuthWaitTime = 2 * time.Second
	systemUserID    = "root"
)

const (
	// AuthLifetime is the duration for which an authorization is considered valid
	AuthLifetime = 15 * time.Minute
	// AuthUserIDKey is the key for the user ID entry in the authorization container
	AuthUserIDKey = "userID"
	// AuthCookieName is the cookie name for cookie-based authentication
	AuthCookieName = "auth"
)

// API represents the admin API in version 1.x
type API struct {
	ctx *service.Context
	log log15.Logger
}

// NewAPI creates a new admin API
func NewAPI(ctx *service.Context) *API {
	a := &API{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1/admin"}),
	}
	return a
}

// GetUserID returns a utility handler. This endpoint displays the user ID, which is stored
// in the authorization container
func (a *API) GetUserID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ctx := service.RequestContext(r)
		userID, ok := ctx.Value(AuthUserIDKey).(string)
		if !ok {
			a.log.Error("internal error. missing userID")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write([]byte(userID))
	})
}
