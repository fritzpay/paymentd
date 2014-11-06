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

func TestPrincipalSQLTx(t *testing.T) {
	Convey("Given a principal DB connection", t, testutil.WithPrincipalDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})
		Convey("Given a transaction", func() {
			tx, err := db.Begin()
			So(err, ShouldBeNil)
			Reset(func() {
				err = tx.Rollback()
				So(err, ShouldBeNil)
			})

			Convey("Given a principal", func() {
				pr := &Principal{
					Created:   time.Now(),
					CreatedBy: "test",
					Name:      "testins",
				}

				Convey("When inserting the principal", func() {
					err = InsertPrincipalTx(tx, pr)

					Convey("It should succeed", func() {
						So(err, ShouldBeNil)
						So(pr.ID, ShouldNotBeEmpty)

						Convey("When selecting the principal by ID", func() {
							selPr, err := PrincipalByIDTx(tx, pr.ID)

							Convey("It should match", func() {
								So(err, ShouldBeNil)
								So(selPr.Name, ShouldEqual, pr.Name)
								So(selPr.Created.Unix(), ShouldEqual, pr.Created.Unix())
								So(selPr.CreatedBy, ShouldEqual, pr.CreatedBy)
							})
						})

						Convey("When selecting the principal by name", func() {
							selPr, err := PrincipalByNameTx(tx, pr.Name)

							Convey("It should match", func() {
								So(err, ShouldBeNil)
								So(selPr.Name, ShouldEqual, pr.Name)
								So(selPr.Created.Unix(), ShouldEqual, pr.Created.Unix())
								So(selPr.CreatedBy, ShouldEqual, pr.CreatedBy)
							})
						})
					})
				})
			})
		})
	}))
}
