package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/api/v1/admin"
	"github.com/fritzpay/paymentd/pkg/service/api/v1/payment"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	servicePath = "/v1"
)

// Service represents the API service version 1.x
type Service struct {
	router *mux.Router

	log log15.Logger
}

// NewService creates a new API service
// It requires a valid service context and takes a router to which
// the service routes will be attached
func NewService(ctx *service.Context, r *mux.Router) (*Service, error) {
	s := &Service{
		router: r.PathPrefix(servicePath).Subrouter(),

		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1"}),
	}

	cfg := ctx.Config()

	if cfg.API.ServeAdmin {
		s.log.Info("registering admin API...")
		admin.NewAPI(s.router, s.log)
	}

	s.log.Info("register payment API...")
	payment.NewAPI(s.router)
	return s, nil
}
