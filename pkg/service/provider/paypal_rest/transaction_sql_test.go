package paypal_rest_test

import (
	"database/sql"
	"testing"

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
		})
	}))
}
