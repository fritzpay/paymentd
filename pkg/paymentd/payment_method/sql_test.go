package payment_method

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/paymentd/provider"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPaymentMethodSQL(t *testing.T) {
	Convey("Given a payment DB connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Convey("Given a principal DB connection", testutil.WithPrincipalDB(t, func(prDB *sql.DB) {

			Convey("Given a test principal", func() {
				princ := principal.Principal{}
				princ.Name = "payment_method_testprincipal"
				princ.CreatedBy = "test"
				err := principal.InsertPrincipalDB(prDB, &princ)
				So(err, ShouldBeNil)
				So(princ.ID, ShouldNotEqual, 0)
				So(princ.Empty(), ShouldBeFalse)

				Reset(func() {
					_, err = prDB.Exec("delete from principal where name = 'payment_method_testprincipal'")
					So(err, ShouldBeNil)
				})

				Convey("Given a test project", func() {
					proj := project.Project{}
					proj.PrincipalID = princ.ID
					proj.Name = "payment_method_testproject"
					proj.CreatedBy = "test"
					err := project.InsertProjectDB(prDB, &proj)
					So(err, ShouldBeNil)
					So(proj.IsValid(), ShouldBeTrue)

					Reset(func() {
						_, err = prDB.Exec("delete from project where name = 'payment_method_testproject'")
						So(err, ShouldBeNil)
					})

					Convey("Given a transaction", func() {
						tx, err := db.Begin()
						So(err, ShouldBeNil)

						Reset(func() {
							err = tx.Rollback()
							So(err, ShouldBeNil)
						})

						Convey("Given a test provider exists", func() {
							pr, err := provider.ProviderByIDTx(tx, 1)
							So(err, ShouldBeNil)
							So(pr.ID, ShouldEqual, 1)

							Convey("When inserting a new payment method", func() {
								pm := PaymentMethod{}
								pm.ProjectID = proj.ID
								pm.Provider.ID = pr.ID
								pm.MethodKey = "test"
								pm.CreatedBy = "test"

								pm.ID, err = InsertPaymentMethodTx(tx, pm)
								So(err, ShouldBeNil)
								So(pm.ID, ShouldNotEqual, 0)
							})
						})
					})
				})
			})
		}))
	}))
}
