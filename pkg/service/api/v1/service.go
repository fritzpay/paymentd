package v1

import (
	"github.com/fritzpay/paymentd/pkg/config"
	"github.com/fritzpay/paymentd/pkg/service/api/v1/admin"
	"github.com/fritzpay/paymentd/pkg/service/api/v1/payment"
	"github.com/gorilla/mux"
)

const (
	servicePath = "/v1"
)

type Service struct {
	router *mux.Router
}

func NewService(cfg config.Config, r *mux.Router) *Service {
	s := &Service{
		router: r.PathPrefix(servicePath).Subrouter(),
	}

	if cfg.API.ServeAdmin {
		admin.NewAPI(s.router)
	}
	payment.NewAPI(s.router)
	return s
}
