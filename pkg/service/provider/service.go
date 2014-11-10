package provider

import (
	"errors"

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

	drivers map[int64]Driver
}

func NewService(ctx *service.Context) (*Service, error) {
	s := &Service{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/provider",
		}),

		drivers: make(map[int64]Driver),
	}
	return s, nil
}

func (s *Service) AttachDrivers(mux *mux.Router) error {
	// add drivers
	s.drivers[driverFritzpay] = &fritzpay.Driver{}

	var err error
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
	if dr, ok := s.drivers[method.Provider.ID]; !ok {
		return nil, ErrNoDriver
	} else {
		return dr, nil
	}
}
