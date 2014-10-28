package v1

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
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
func NewService(ctx *service.Context, mux *http.ServeMux) (*Service, error) {
	s := &Service{
		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1"}),
	}

	cfg := ctx.Config()

	if cfg.API.ServeAdmin {
		s.log.Info("registering admin API...")

		admin := NewAdminAPI(ctx)
		mux.Handle(ServicePath+"/authorization", admin.AuthorizationHandler())
		mux.Handle(ServicePath+"/authorization/", admin.AuthorizeHandler())
		mux.Handle(ServicePath+"/user", admin.AuthRequiredHandler(admin.GetUserID()))

		mux.Handle(ServicePath+"/principal", admin.AuthRequiredHandler(admin.PrincipalRequest()))
		mux.Handle(ServicePath+"/principal/", admin.AuthRequiredHandler(admin.PrincipalGetRequest()))
		mux.Handle(ServicePath+"/provider", admin.AuthRequiredHandler(admin.ProviderGetAllRequest()))
		mux.Handle(ServicePath+"/provider/", admin.AuthRequiredHandler(admin.ProviderGetRequest()))
		mux.Handle(ServicePath+"/project", admin.AuthRequiredHandler(admin.ProjectRequest()))
		mux.Handle(ServicePath+"/project/", admin.AuthRequiredHandler(admin.ProjectGetRequest()))
		mux.Handle(ServicePath+"/currency", admin.AuthRequiredHandler(admin.CurrencyGetAllRequest()))
		mux.Handle(ServicePath+"/currency/", admin.AuthRequiredHandler(admin.CurrencyGetRequest()))
	}

	s.log.Info("initializing payment ID encoder...")
	idCoder, err := payment.NewIDEncoder(cfg.Payment.PaymentIDEncPrime, cfg.Payment.PaymentIDEncXOR)
	if err != nil {
		s.log.Error("error initializing payment ID encoder", log15.Ctx{"err": err})
		return nil, err
	}
	ctx = ctx.WithValue(serviceContextPaymentIDEncoder, idCoder)

	s.log.Info("registering payment API...")
	payment := NewPaymentAPI(ctx)
	mux.Handle(ServicePath+"/payment", payment.InitPayment())

	return s, nil
}
