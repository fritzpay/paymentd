package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"time"
)

const (
	systemUserID = "root"
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
type AdminAPI struct {
	ctx *service.Context
	log log15.Logger
}

// type used for formated AdminAPI Responses
type AdminAPIResponse struct {
	ServiceResponse
}

var (
	ErrConflict = ServiceResponse{
		http.StatusConflict,
		StatusError,
		"resource already exits",
		nil,
		"resource already exits",
	}
	ErrReadParam = ServiceResponse{
		http.StatusBadRequest,
		StatusError,
		"parameter malformed",
		nil,
		"parameter malformed",
	}
	ErrMethod = ServiceResponse{
		http.StatusMethodNotAllowed,
		StatusError,
		"method not allowed",
		nil,
		"method not allowed",
	}
)

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
