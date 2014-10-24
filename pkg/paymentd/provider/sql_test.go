package provider

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestProviderSQL(t *testing.T) {
	Convey("Given a payment DB connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})
		Convey("When selecting the test provider", func() {
			pr, err := ProviderByIDDB(db, 1)

			Convey("It should return the test provider", func() {
				So(err, ShouldBeNil)
				So(pr.ID, ShouldEqual, 1)
				So(pr.Name, ShouldEqual, "fritzpay")
			})
		})

		Convey("When selecting a nonexistent provider", func() {
			pr, err := ProviderByIDDB(db, 0)

			Convey("It should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrProviderNotFound)
				So(pr.ID, ShouldEqual, 0)
			})
		})
	}))
}
