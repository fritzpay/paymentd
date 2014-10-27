package payment_test

import (
	"database/sql"
	. "github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestPaymentTokenGenerationSQL(t *testing.T) {
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
						Convey("When generating a token for the payment", func() {
							t, err := CreatePaymentToken(p.PaymentID())
							So(err, ShouldBeNil)
							So(t.Valid(time.Minute), ShouldBeTrue)

							Convey("When inserting the token", func() {
								err = InsertPaymentTokenTx(tx, &t)

								Convey("It should succeed", func() {
									So(err, ShouldBeNil)

									Convey("Given a duplicate token", func() {
										t2 := t
										So(t.Token, ShouldEqual, t2.Token)
										Convey("When inserting a duplicate token", func() {
											err = InsertPaymentTokenTx(tx, &t2)
											Convey("It should succeed", func() {
												So(err, ShouldBeNil)
											})
											Convey("It should have regenerated the token", func() {
												So(t.Token, ShouldNotEqual, t2.Token)
											})
										})
									})
								})
							})
						})
					}))
				})
			}))
		}))
	}))
}
