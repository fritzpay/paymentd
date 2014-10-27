package payment_test

import (
	"database/sql"
	. "github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func WithTestProject(db, prDB *sql.DB, f func(pr project.Project)) func() {
	return func() {
		princ := principal.Principal{}
		princ.Name = "payment_testprincipal"
		princ.CreatedBy = "test"
		err := principal.InsertPrincipalDB(prDB, &princ)
		So(err, ShouldBeNil)
		So(princ.ID, ShouldNotEqual, 0)
		So(princ.Empty(), ShouldBeFalse)

		proj := project.Project{}
		proj.PrincipalID = princ.ID
		proj.Name = "payment_testproject"
		proj.CreatedBy = "test"
		err = project.InsertProjectDB(prDB, &proj)
		So(err, ShouldBeNil)
		So(proj.IsValid(), ShouldBeTrue)

		Reset(func() {
			_, err = prDB.Exec("delete from project where name = 'payment_testproject'")
			So(err, ShouldBeNil)
			_, err = prDB.Exec("delete from principal where name = 'payment_testprincipal'")
			So(err, ShouldBeNil)
		})

		f(proj)
	}
}

func WithTestPayment(tx *sql.Tx, pr project.Project, f func(p Payment)) func() {
	return func() {
		p := &Payment{}
		err := p.SetProject(pr)
		So(err, ShouldBeNil)

		p.Amount = 1234
		p.Subunits = 2
		p.Currency = "EUR"
		p.Created = time.Now()

		err = InsertPaymentTx(tx, p)
		So(err, ShouldBeNil)

		f(*p)
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
