// +build ignore

// ignored since go-sqlmock does not yet support instances
package currency

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestProjectSQLMapping(t *testing.T) {
	Convey("Given a database mock connection", t, func() {
		mockID, mock, err := sqlmock.NewMockConn()
		So(err, ShouldBeNil)
		mock.ExpectQuery("SELECT(.+)FROM currency(.+)code_iso_4217 = ?").
			WithArgs("nonexistent").
			WillReturnRows(sqlmock.NewRows(string("EUR")))

		db, err := sql.Open("mock", "code_iso_4217="+mockID)
		So(err, ShouldBeNil)
		err = db.Ping()
		So(err, ShouldBeNil)

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

		Convey("When requesting all currencies", func() {
			currencyList, err CurrencyAllDB(db)

			Convey("It should return a list of currencies", func (){

				})
		})

	})

}
