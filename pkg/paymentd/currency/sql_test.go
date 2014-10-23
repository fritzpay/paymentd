package currency

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCurrencySQL(t *testing.T) {
	Convey("Given a payment DB connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {

		Convey("When requesting a nonexistent currency", func() {
			currency, err := CurrencyByCodeISO4217DB(db, "nonexistent")

			Convey("It should return an empty currency", func() {
				So(currency.IsEmpty(), ShouldBeTrue)
				Convey("The mock should complete successfully", func() {
					err = db.Close()
					So(err, ShouldBeNil)
				})
			})
			Convey("It should return an error not found", func() {
				So(err, ShouldEqual, ErrCurrencyNotFound)
				Convey("The mock should complete successfully", func() {
					err = db.Close()
					So(err, ShouldBeNil)
				})
			})
		})
	}))
}
