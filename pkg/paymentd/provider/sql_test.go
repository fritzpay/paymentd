package provider

import (
	"database/sql"
	"testing"

	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProviderSQL(t *testing.T) {
	Convey("Given a payment DB connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})
		Convey("When selecting the test provider", func() {
			pr, err := ProviderByNameDB(db, "fritzpay")

			Convey("It should return the test provider", func() {
				So(err, ShouldBeNil)
				So(pr.Name, ShouldEqual, "fritzpay")
			})
		})

		Convey("When selecting a nonexistent provider", func() {
			pr, err := ProviderByNameDB(db, "0")

			Convey("It should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrProviderNotFound)
				So(pr.Name, ShouldEqual, "")
			})
		})
	}))
}
