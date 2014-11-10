package payment_test

import (
	"database/sql"
	"testing"
	"time"

	. "github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	. "github.com/smartystreets/goconvey/convey"
)

func WithTestProject(db, prDB *sql.DB, f func(pr *project.Project)) func() {
	return func() {
		princ, err := principal.PrincipalByNameDB(prDB, "testprincipal")
		So(err, ShouldBeNil)
		So(princ.ID, ShouldNotEqual, 0)
		So(princ.Empty(), ShouldBeFalse)

		proj, err := project.ProjectByPrincipalIDNameDB(prDB, princ.ID, "testproject")
		So(err, ShouldBeNil)

		f(proj)
	}
}

func WithTestPayment(tx *sql.Tx, pr *project.Project, f func(p *Payment)) func() {
	return func() {
		p := &Payment{}
		err := p.SetProject(pr)
		So(err, ShouldBeNil)

		p.Amount = 1234
		p.Subunits = 2
		p.Currency = "EUR"
		p.Created = time.Unix(1234, 0)

		err = InsertPaymentTx(tx, p)
		So(err, ShouldBeNil)

		f(p)
	}
}

func TestPaymentAmountDecimal(t *testing.T) {
	Convey("Given a payment", t, func() {
		p := &Payment{}
		p.Amount = 1234
		p.Subunits = 2

		Convey("When retrieving the decimal amount representation", func() {
			dec := p.Decimal()

			Convey("It should be correctly represented", func() {
				So(dec.String(), ShouldEqual, "12.34")
			})
		})
	})
}
