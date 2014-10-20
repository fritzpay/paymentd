package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

const (
	serviceVersion = "v1"
	// ServicePath is the (sub-)path under which the API service v1.x resides in
	ServicePath = "/" + serviceVersion
)

// Service represents the API service version 1.x
type Service struct {
	log log15.Logger
}

// NewService creates a new API service
// It requires a valid service context and takes a router to which
// the service routes will be attached
func NewService(ctx *service.Context, mux *http.ServeMux) *Service {
	s := &Service{
		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1"}),
	}

	cfg := ctx.Config()

	if cfg.API.ServeAdmin {
		s.log.Info("registering admin API...")

		admin := NewAdminAPI(ctx)
		mux.Handle(ServicePath+"/authorization/", admin.GetAuthorization())
		mux.Handle(ServicePath+"/user/", admin.AuthRequiredHandler(admin.GetUserID()))

		mux.Handle(ServicePath+"/principal/", admin.AuthRequiredHandler(admin.PrincipalRequest()))
	}

	s.log.Info("registering payment API...")
	NewPaymentAPI(ctx)

	return s
}
