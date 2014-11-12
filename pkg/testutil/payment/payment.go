package payment

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	. "github.com/smartystreets/goconvey/convey"
)

func WithPaymentInTx(tx *sql.Tx, f func(p *payment.Payment)) func() {
	return func() {
		pr := &project.Project{
			ID:          1,
			PrincipalID: 1,
		}
		method, err := payment_method.PaymentMethodByIDTx(tx, 1)
		So(err, ShouldBeNil)
		So(method.Active(), ShouldBeTrue)
		So(method.ID, ShouldNotEqual, 0)

		err = payment_method.InsertPaymentMethodStatusTx(tx, method)
		So(err, ShouldBeNil)

		p := &payment.Payment{
			Created:  time.Now(),
			Ident:    "test_" + fmt.Sprintf("%d", time.Now().UnixNano()),
			Amount:   1234,
			Subunits: 2,
			Currency: "EUR",
		}
		err = p.SetProject(pr)
		So(err, ShouldBeNil)
		p.Config.SetCountry("DE")
		p.Config.SetLocale("en-US")
		p.Config.SetPaymentMethodID(method.ID)

		err = payment.InsertPaymentTx(tx, p)
		So(err, ShouldBeNil)
		err = payment.InsertPaymentConfigTx(tx, p)
		So(err, ShouldBeNil)

		p, err = payment.PaymentByIDTx(tx, p.PaymentID())
		So(err, ShouldBeNil)

		f(p)
	}
}
