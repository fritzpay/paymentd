package admin

import (
	"gopkg.in/inconshreveable/log15.v2"
)

// API represents the admin API in version 1.x
type API struct {
	log log15.Logger
}

// NewAPI creates a new admin API
func NewAPI(log log15.Logger) *API {
	a := &API{
		log: log.New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1/admin"}),
	}
	return a
}
