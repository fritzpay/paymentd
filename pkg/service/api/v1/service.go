package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
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
func NewService(ctx *service.Context, mux *mux.Router) *Service {
	s := &Service{
		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1"}),
	}

	cfg := ctx.Config()

	if cfg.API.ServeAdmin {
		s.log.Info("registering admin API...")

		admin := NewAdminAPI(ctx)
		mux.Handle(ServicePath+"/authorization", admin.AuthorizationHandler())
		mux.Handle(ServicePath+"/authorization/{method}", admin.AuthorizeHandler())
		mux.Handle(ServicePath+"/user", admin.AuthRequiredHandler(admin.GetUserID()))

		mux.Handle(ServicePath+"/principal", admin.AuthRequiredHandler(admin.PrincipalRequest()))
		mux.Handle(ServicePath+"/principal/{name}", admin.AuthRequiredHandler(admin.PrincipalGetRequest()))
		mux.Handle(ServicePath+"/provider", admin.AuthRequiredHandler(admin.ProviderGetAllRequest()))
		mux.Handle(ServicePath+"/provider/{id}", admin.AuthRequiredHandler(admin.ProviderGetRequest()))
		mux.Handle(ServicePath+"/project", admin.AuthRequiredHandler(admin.ProjectRequest()))
		mux.Handle(ServicePath+"/project/{projectid}", admin.AuthRequiredHandler(admin.ProjectGetRequest()))
		mux.Handle(ServicePath+"/project/{projectid}/method", admin.AuthRequiredHandler(admin.PaymentMethodsGetRequest()))
		mux.Handle(ServicePath+"/project/{projectid}/method/{methodid}", admin.AuthRequiredHandler(admin.PaymentMethodsRequest()))
		mux.Handle(ServicePath+"/currency", admin.AuthRequiredHandler(admin.CurrencyGetAllRequest()))
		mux.Handle(ServicePath+"/currency/{currencycode}", admin.AuthRequiredHandler(admin.CurrencyGetRequest()))
	}

	s.log.Info("registering payment API...")
	payment := NewPaymentAPI(ctx)
	mux.Handle(ServicePath+"/payment", payment.InitPayment())

	return s
}
