package principal

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPrincipalSQL(t *testing.T) {
	Convey("Given a principal DB connection", t, testutil.WithPrincipalDB(t, func(db *sql.DB) {
		Convey("When requesting a nonexistent principal", func() {
			principal, err := PrincipalByNameDB(db, "nonexistent")

			Convey("It should return an empty principal", func() {
				So(principal.Empty(), ShouldBeTrue)
			})
			Convey("It should return an error not found", func() {
				So(err, ShouldEqual, ErrPrincipalNotFound)
			})
		})

	}))

}
