package payment

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/nonce"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/payment/notification"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"time"
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
	projectKey, err := project.ProjectKeyByKeyDB(s.ctx.PrincipalDB(service.ReadOnly), cbProjectKey)
	if err != nil {
		if err == project.ErrProjectKeyNotFound {
			log.Error("invalid project key")
			return
		}
		log.Error("error retrieving project key", log15.Ctx{"err": err})
		return
	}
	if !projectKey.IsValid() {
		log.Warn("cannot notify with invalid project key", log15.Ctx{"projectKey": projectKey})
		return
	}
	// balance
	tl, err := payment.PaymentTransactionsBeforeDB(s.ctx.PaymentDB(service.ReadOnly), paymentTx)
	if err != nil && err != payment.ErrPaymentTransactionNotFound {
		log.Error("error retrieving transaction history", log15.Ctx{"err": err})
		return
	}
	if tl != nil {
		not.SetTransactions(tl)
	}
	// signing
	non, err := nonce.New()
	if err != nil {
		log.Error("error generating nonce", log15.Ctx{"err": err})
		return
	}
	secret, err := projectKey.SecretBytes()
	if err != nil {
		log.Error("error retrieving secret", log15.Ctx{"err": err})
		return
	}
	err = not.Sign(time.Now(), non.Nonce, secret)
	if err != nil {
		log.Error("error signing notification", log15.Ctx{"err": err})
		return
	}

	req, err := http.NewRequest("POST", cbURL, not.Reader())
	if err != nil {
		log.Error("error creating HTTP request", log15.Ctx{"err": err})
		return
	}
	req.Header.Set("User-Agent", not.Identification())
	req.Close = true
	res, err := s.cl.Do(req)
	if err != nil {
		log.Error("error on HTTP request", log15.Ctx{"err": err})
	} else {
		log.Info("notified", log15.Ctx{"HTTPStatusCode": res.StatusCode})
	}
}
