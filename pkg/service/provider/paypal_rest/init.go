package paypal_rest

import (
	"database/sql"
	"net/http"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
)

func (d *Driver) InitPayment(p *payment.Payment, method *payment_method.Method) (http.Handler, error) {
	log := d.log.New(log15.Ctx{
		"method":          "InitPayment",
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
		return nil, ErrDatabase
	}
	return nil, nil
}
