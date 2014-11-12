package paypal_rest

import (
	"database/sql"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
)

func (d *Driver) InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error) {

	return d.InitPageHandler(p), nil
}

func (d *Driver) doInit(errors chan<- error, p *payment.Payment, method *payment_method.Method) {
	log := d.log.New(log15.Ctx{
		"method":          "doInit",
		"projectID":       p.ProjectID(),
		"paymentID":       p.ID(),
		"paymentMethodID": method.ID,
	})
	if Debug {
		log.Debug("initializing payment...")
	}

	var tx *sql.Tx
	var err error
	var commit bool
	defer func() {
		if tx != nil && !commit {
			err = tx.Rollback()
			if err != nil {
				log.Crit("error on rollback", log15.Ctx{"err": err})
			}
		}
	}()
	tx, err = d.ctx.PaymentDB().Begin()
	if err != nil {
		commit = true
		log.Crit("error on begin tx", log15.Ctx{"err": err})
		errors <- ErrDatabase
		return
	}

	tr, err := d.oAuthTransport(log)(tx, p, method)
	if err != nil {
		errors <- err
		return
	}
	err = tr.AuthenticateClient()
	if err != nil {
		log.Error("error authenticating", log15.Ctx{"err": err})
		errors <- ErrInternal
		return
	}
	if Debug {
		log.Debug("authenticated", log15.Ctx{"token": tr.AccessToken})
	}

	commit = true
	err = tx.Commit()
	if err != nil {
		log.Crit("error on commit", log15.Ctx{"err": err})
		errors <- ErrDatabase
		return
	}
	close(errors)
}
