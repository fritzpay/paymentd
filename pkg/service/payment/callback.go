package payment

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/service/payment/notification"
	"gopkg.in/inconshreveable/log15.v2"
)

// Callbacker describes a type that can provide information about callbacks to be made
type Callbacker interface {
	HasCallback() bool
	CallbackConfig() (url, apiVersion, projectKey string)
}

func CanCallback(c Callbacker) bool {
	return c.HasCallback()
}

func (s *Service) Notify(c Callbacker, paymentTx *payment.PaymentTransaction) {
	cbURL, cbAPIVersion, cbProjectKey := c.CallbackConfig()
	log := s.log.New(log15.Ctx{
		"method":                      "Notify",
		"projectID":                   paymentTx.Payment.ProjectID(),
		"paymentID":                   paymentTx.Payment.ID(),
		"paymentTransactionTimestamp": paymentTx.Timestamp.UnixNano(),
		"callbackURL":                 cbURL,
		"callbackAPIVersion":          cbAPIVersion,
		"callbackProjectKey":          cbProjectKey,
	})
	log.Info("notifying...")
	notF, err := notification.NotificationByVersion(cbAPIVersion)
	if err != nil {
		log.Error("error retrieving notification by version", log15.Ctx{"err": err})
		return
	}
	not, err := notF(s.EncodedPaymentID(paymentTx.Payment.PaymentID()), paymentTx.Payment)
	if err != nil {
		log.Error("error creating notification", log15.Ctx{"err": err})
		return
	}
	_ = not
}
