package v1

import (
	"time"

	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	systemUserID = "root"
)

const (
	// AuthLifetime is the duration for which an authorization is considered valid
	// TODO @configure
	AuthLifetime = 15 * time.Minute
	// AuthUserIDKey is the key for the user ID entry in the authorization container
	AuthUserIDKey = "userID"
	// AuthCookieName is the cookie name for cookie-based authentication
	AuthCookieName = "auth"
)

// API represents the admin API in version 1.x
type AdminAPI struct {
	ctx *service.Context
	log log15.Logger
}

// type used for formated AdminAPI Responses
type AdminAPIResponse struct {
	ServiceResponse
}

// NewAPI creates a new admin API
func NewAdminAPI(ctx *service.Context) *AdminAPI {
	a := &AdminAPI{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1",
			"API": "AdminAPI",
		}),
	}
	return a
}
