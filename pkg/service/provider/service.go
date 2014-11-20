package provider

import (
	"errors"

	"github.com/fritzpay/paymentd/pkg/paymentd/provider"

	"github.com/fritzpay/paymentd/pkg/service/provider/paypal_rest"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/provider/fritzpay"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	ProviderPath = "/p"
)

var (
	ErrNoDriver = errors.New("no driver found")
)

type Service struct {
	ctx *service.Context
	log log15.Logger

	drivers map[string]Driver
}

func NewService(ctx *service.Context) (*Service, error) {
	s := &Service{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/provider",
		}),

		drivers: make(map[string]Driver),
	}
	return s, nil
}

func (s *Service) AttachDrivers(mux *mux.Router) error {
	providers, err := provider.ProviderAllDB(s.ctx.PaymentDB())
	if err != nil {
		s.log.Error("error retrieving providers", log15.Ctx{"err": err})
		return err
	}
	// add drivers
	for _, prov := range providers {
		s.log.Info("attaching provider driver...", log15.Ctx{
			"providerName": prov.Name,
		})
		switch prov.Name {
		case driverFritzpay:
			s.drivers[driverFritzpay] = &fritzpay.Driver{}
		case driverPaypalREST:
			s.drivers[driverPaypalREST] = &paypal_rest.Driver{}
		default:
			s.log.Error("unknown provider id in database", log15.Ctx{"providerName": prov.Name})
			return ErrNoDriver
		}
	}

	mux = mux.PathPrefix(ProviderPath).Subrouter()
	for _, dr := range s.drivers {
		err = dr.Attach(s.ctx, mux)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) Driver(method *payment_method.Method) (Driver, error) {
	if dr, ok := s.drivers[method.Provider.Name]; !ok {
		return nil, ErrNoDriver
	} else {
		return dr, nil
	}
}
