package v1

import (
	"code.google.com/p/go.net/context"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/config"
	"github.com/fritzpay/paymentd/pkg/service/api/v1/admin"
	"github.com/fritzpay/paymentd/pkg/service/api/v1/payment"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	servicePath = "/v1"
)

type Service struct {
	router *mux.Router

	log log15.Logger
}

func NewService(ctx context.Context) (*Service, error) {
	var r *mux.Router
	var log log15.Logger
	var cfg config.Config
	var ok bool

	if r, ok = ctx.Value("router").(*mux.Router); !ok {
		return nil, fmt.Errorf("invalid context. require router, got %T", ctx.Value("router"))
	}
	if log, ok = ctx.Value("log").(log15.Logger); !ok {
		return nil, fmt.Errorf("invalid context. require log, got %T", ctx.Value("log"))
	}

	s := &Service{
		router: r.PathPrefix(servicePath).Subrouter(),

		log: log.New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1"}),
	}

	if cfg, ok = ctx.Value("cfg").(config.Config); !ok {
		s.log.Crit("invalid context. require config", log15.Ctx{
			"cfgType": fmt.Sprintf("%T", ctx.Value("cfg")),
		})
		return nil, fmt.Errorf("invalid context. require config, got %T", ctx.Value("cfg"))
	}

	if cfg.API.ServeAdmin {
		s.log.Info("registering admin API...")
		admin.NewAPI(s.router, s.log)
	}

	s.log.Info("register payment API...")
	payment.NewAPI(s.router)
	return s, nil
}
