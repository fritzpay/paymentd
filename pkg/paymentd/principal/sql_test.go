package principal

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func WithPrincipal(db *sql.DB, f func(pr *Principal)) func() {
	return func() {
		pr := &Principal{
			Created:   time.Now(),
			CreatedBy: "test",
			Name:      "test_principal",
		}
		err := InsertPrincipalDB(db, pr)
		So(err, ShouldBeNil)
		So(pr.Empty(), ShouldBeFalse)

		Reset(func() {
			_, err := db.Exec("delete from principal where name = 'test_principal'")
			So(err, ShouldBeNil)
		})

		f(pr)
	}
}

func TestPrincipalSQL(t *testing.T) {
	Convey("Given a principal DB connection", t, testutil.WithPrincipalDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})
		Convey("When requesting a nonexistent principal", func() {
			principal, err := PrincipalByNameDB(db, "nonexistent")

			Convey("It should return an empty principal", func() {
				So(principal.Empty(), ShouldBeTrue)
			})
			Convey("It should return an error not found", func() {
				So(err, ShouldEqual, ErrPrincipalNotFound)
			})
		})

		Convey("Given a principal", WithPrincipal(db, func(pr *Principal) {

			Convey("When selecting a principal by name", func() {
				selPr, err := PrincipalByNameDB(db, pr.Name)

				Convey("It should succeed", func() {
					So(err, ShouldBeNil)
					So(selPr.Empty(), ShouldBeFalse)
					Convey("It should match", func() {
						So(selPr.ID, ShouldEqual, pr.ID)
					})
				})
			})
		}))
	}))
}
