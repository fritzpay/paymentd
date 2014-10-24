package payment_test

import (
	"database/sql"
	. "github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPaymentSQL(t *testing.T) {
	Convey("Given a payment DB", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})
		Convey("Given a principal DB", testutil.WithPrincipalDB(t, func(prDB *sql.DB) {
			Reset(func() {
				prDB.Close()
			})
			Convey("Given a test project", WithTestProject(db, prDB, func(proj project.Project) {
				Convey("Given a transaction", func() {
					tx, err := db.Begin()
					So(err, ShouldBeNil)

					Reset(func() {
						err = tx.Rollback()
						So(err, ShouldBeNil)
					})

					Convey("Given a test payment", WithTestPayment(tx, proj, func(p Payment) {
						Convey("When selecting a payment by ident", func() {
							p2, err := PaymentByProjectIDAndIdentTx(tx, proj.ID, p.Ident)
							Convey("It should succeed", func() {
								So(err, ShouldBeNil)
								Convey("It should match the original payment", func() {
									So(p2.ID(), ShouldEqual, p.ID())
								})
							})
						})
					}))
				})
			}))
		}))
	}))
}
