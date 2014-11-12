package paypal_rest_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"

	"github.com/fritzpay/paymentd/pkg/service/provider/paypal_rest"

	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPaypalTransaction(t *testing.T) {
	Convey("Given a payment DB", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Convey("Given a db tx", func() {
			tx, err := db.Begin()
			So(err, ShouldBeNil)
			Reset(func() {
				err = tx.Rollback()
				So(err, ShouldBeNil)
			})

			Convey("Given a payment", testutil.WithPaymentInTx(tx, func(p *payment.Payment) {

				Convey("When creating a new transaction", func() {
					pt := &paypal_rest.Transaction{
						ProjectID: p.ProjectID(),
						PaymentID: p.ID(),
						Timestamp: time.Now(),
						Type:      "test",
					}

					Convey("When no paypal update time is set", func() {
						So(pt.PaypalUpdateTime, ShouldBeNil)

						Convey("When inserting the transaction", func() {
							err = paypal_rest.InsertTransaction(tx, pt)

							Convey("It should succeed", func() {
								So(err, ShouldBeNil)

								Convey("When retrieving the inserted transaction", func() {
									ptRet, err := paypal_rest.TransactionCurrentByPaymentIDTx(tx, p.PaymentID())

									Convey("It should succeed", func() {
										So(err, ShouldBeNil)
										So(ptRet, ShouldNotBeNil)
										Convey("It should have no update time", func() {
											So(ptRet.PaypalUpdateTime, ShouldBeNil)
										})
									})
								})
							})
						})
					})

					Convey("When an update time is set", func() {
						tm := time.Now()
						pt.PaypalUpdateTime = &tm

						Convey("When inserting the transaction", func() {
							err = paypal_rest.InsertTransaction(tx, pt)

							Convey("It should succeed", func() {
								So(err, ShouldBeNil)

								Convey("When retrieving the inserted transaction", func() {
									ptRet, err := paypal_rest.TransactionCurrentByPaymentIDTx(tx, p.PaymentID())

									Convey("It should succeed", func() {
										So(err, ShouldBeNil)
										So(ptRet, ShouldNotBeNil)
										Convey("It should have an update time", func() {
											So(ptRet.PaypalUpdateTime, ShouldNotBeNil)
										})
									})
								})
							})
						})
					})
				})
			}))
		})
	}))
}
