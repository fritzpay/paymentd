package provider

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	ProviderPath = "/p"
)

type Service struct {
	ctx *service.Context
	log log15.Logger
}

func NewService(ctx *service.Context) (*Service, error) {
	s := &Service{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/provider",
		}),
	}
	return s, nil
}

func (s *Service) AttachDrivers(mux *mux.Router) {
	mux = mux.PathPrefix(ProviderPath).Subrouter()
	for _, dr := range drivers {
		dr.Attach(s.ctx, mux)
	}
}
