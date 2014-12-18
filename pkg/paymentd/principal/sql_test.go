package principal

import (
	"database/sql"
	"testing"
	"time"

	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
)

func WithPrincipal(db *sql.Tx, f func(pr *Principal)) func() {
	return func() {
		pr := &Principal{
			Created:   time.Now(),
			CreatedBy: "test",
			Name:      "test_principal",
			Status:    PrincipalStatusActive,
		}
		err := InsertPrincipalTx(db, pr)
		So(err, ShouldBeNil)
		So(pr.Empty(), ShouldBeFalse)
		err = InsertPrincipalStatusTx(db, *pr, "test2")
		So(err, ShouldBeNil)

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

		Convey("Given a db transaction", func() {
			tx, err := db.Begin()
			So(err, ShouldBeNil)
			Reset(func() {
				err = tx.Rollback()
				So(err, ShouldBeNil)
			})

			Convey("Given a principal", WithPrincipal(tx, func(pr *Principal) {

				Convey("When selecting a principal by name", func() {
					selPr, err := PrincipalByNameTx(tx, pr.Name)

					Convey("It should succeed", func() {
						So(err, ShouldBeNil)
						So(selPr.Empty(), ShouldBeFalse)
						Convey("It should match", func() {
							So(selPr.ID, ShouldEqual, pr.ID)
						})
					})
				})
			}))
		})
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

						Convey("When selecting the principal without status", func() {
							selPr, err := PrincipalByIDTx(tx, pr.ID)

							Convey("It should return no principal", func() {
								So(selPr.ID, ShouldEqual, 0)
								So(err, ShouldEqual, ErrPrincipalNotFound)
							})
						})

						Convey("When setting the status to active", func() {
							pr.Status = PrincipalStatusActive
							err = InsertPrincipalStatusTx(tx, *pr, "tester")
							So(err, ShouldBeNil)

							Convey("When selecting the principal by ID", func() {
								selPr, err := PrincipalByIDTx(tx, pr.ID)

								Convey("It should match", func() {
									So(err, ShouldBeNil)
									So(selPr.Name, ShouldEqual, pr.Name)
									So(selPr.Created.Unix(), ShouldEqual, pr.Created.Unix())
									So(selPr.CreatedBy, ShouldEqual, pr.CreatedBy)
									So(selPr.Active(), ShouldBeNil)
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

						Convey("When setting the status to deleted", func() {
							pr.Status = PrincipalStatusDeleted
							err = InsertPrincipalStatusTx(tx, *pr, "test3")
							So(err, ShouldBeNil)

							Convey("When selecting the principal", func() {
								selPr, err := PrincipalByIDTx(tx, pr.ID)
								Convey("It should not be found", func() {
									So(selPr.ID, ShouldEqual, 0)
									So(err, ShouldEqual, ErrPrincipalNotFound)
								})
							})
						})
					})
				})
			})
		})
	}))
}
