// +build ignore

// ignored since go-sqlmock does not yet support instances
package principal

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPrincipalSQLMapping(t *testing.T) {
	Convey("Given a database mock connection", t, func() {
		mockID, mock, err := sqlmock.NewMockConn()
		So(err, ShouldBeNil)
		mock.ExpectQuery("SELECT(.+)FROM principal(.+)name = ?").
			WithArgs("nonexistent").
			WillReturnRows(sqlmock.NewRows([]string{"test"}))

		db, err := sql.Open("mock", "id="+mockID)
		So(err, ShouldBeNil)
		err = db.Ping()
		So(err, ShouldBeNil)

		Convey("When requesting a nonexistent principal", func() {
			principal, err := PrincipalByNameDB(db, "nonexistent")

			Convey("It should return an empty principal", func() {
				So(principal.Empty(), ShouldBeTrue)
				Convey("The mock should complete successfully", func() {
					err = db.Close()
					So(err, ShouldBeNil)
				})
			})
			Convey("It should return an error not found", func() {
				So(err, ShouldEqual, ErrPrincipalNotFound)
				Convey("The mock should complete successfully", func() {
					err = db.Close()
					So(err, ShouldBeNil)
				})
			})
		})

	})

}
