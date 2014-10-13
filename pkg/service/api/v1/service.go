package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/api/v1/admin"
	"github.com/fritzpay/paymentd/pkg/service/api/v1/payment"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

const (
	// ServicePath is the (sub-)path under which the API service v1.x resides in
	ServicePath = "/v1/"
)

// Service represents the API service version 1.x
type Service struct {
	log log15.Logger
}

// NewService creates a new API service
// It requires a valid service context and takes a router to which
// the service routes will be attached
func NewService(ctx *service.Context) *Service {
	s := &Service{
		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1"}),
	}

	cfg := ctx.Config()

	if cfg.API.ServeAdmin {
		s.log.Info("registering admin API...")
		admin.NewAPI(ctx)
	}

	s.log.Info("registering payment API...")
	payment.NewAPI()
	return s
}

// ServeHTTP implements the http.Handler and serves the API v1.x
func (s *Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
