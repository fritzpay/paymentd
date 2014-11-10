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

// context keys
const (
	serviceContextPaymentIDEncoder = "PaymentIDEncoder"
)

// Log is the default logger for the API service v1
var Log log15.Logger

func init() {
	Log = log15.New(log15.Ctx{
		"pkg":  "github.com/fritzpay/paymentd/pkg/service/api/v1",
		"type": "ServiceResponse",
	})
	Log.SetHandler(log15.StderrHandler)
}

// Service represents the API service version 1.x
type Service struct {
	log log15.Logger
}

// NewService creates a new API service
// It requires a valid service context and takes a router to which
// the service routes will be attached
func NewService(ctx *service.Context, mux *mux.Router) (*Service, error) {
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
		mux.Handle(ServicePath+"/principal/{name:[-A-Za-z0-9_]+}", admin.AuthRequiredHandler(admin.PrincipalGetRequest()))
		mux.Handle(ServicePath+"/provider", admin.AuthRequiredHandler(admin.ProviderGetAllRequest()))
		mux.Handle(ServicePath+"/provider/{providerid}", admin.AuthRequiredHandler(admin.ProviderGetRequest()))
		mux.Handle(ServicePath+"/project", admin.AuthRequiredHandler(admin.ProjectRequest()))
		mux.Handle(ServicePath+"/project/{projectid}", admin.AuthRequiredHandler(admin.ProjectGetRequest()))
		mux.Handle(ServicePath+"/project/{projectid}/method", admin.AuthRequiredHandler(admin.PaymentMethodsGetRequest()))
		mux.Handle(ServicePath+"/project/{projectid}/method/", admin.AuthRequiredHandler(admin.PaymentMethodsRequest()))
		mux.Handle(ServicePath+"/currency", admin.AuthRequiredHandler(admin.CurrencyGetAllRequest()))
		mux.Handle(ServicePath+"/currency/{currencycode}", admin.AuthRequiredHandler(admin.CurrencyGetRequest()))
	}

	s.log.Info("registering payment API...")
	payment, err := NewPaymentAPI(ctx)
	if err != nil {
		s.log.Error("error registering payment API", log15.Ctx{"err": err})
		return nil, err
	}
	mux.Handle(ServicePath+"/payment", payment.InitPayment()).Methods("POST")
	mux.Handle(ServicePath+"/payment/paymentId/{paymentId}", payment.GetPayment()).Methods("GET")
	mux.Handle(ServicePath+"/payment/PaymentId/{paymentId}", payment.GetPayment()).Methods("GET")
	mux.Handle(ServicePath+"/payment/ident/{ident}", payment.GetPayment()).Methods("GET")
	mux.Handle(ServicePath+"/payment/Ident/{ident}", payment.GetPayment()).Methods("GET")

	return s, nil
}
