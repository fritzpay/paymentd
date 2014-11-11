// +build ignore

// ignored since go-sqlmock does not yet support instances
package project

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestProjectSQLMapping(t *testing.T) {
	Convey("Given a database mock connection", t, func() {
		mockID, mock, err := sqlmock.NewMockConn()
		So(err, ShouldBeNil)
		mock.ExpectQuery("SELECT(.+)FROM project(.+)id = ?").
			WithArgs("nonexistent").
			WillReturnRows(sqlmock.NewRows(int64{1}))

		db, err := sql.Open("mock", "id="+mockID)
		So(err, ShouldBeNil)
		err = db.Ping()
		So(err, ShouldBeNil)

		Convey("When requesting a nonexistent project", func() {
			project, err := ProjectByPrincipalIDNameDB(db, 1, "nonexistent")

			Convey("It should return an empty project", func() {
				So(project.Empty(), ShouldBeTrue)
				Convey("The mock should complete successfully", func() {
					err = db.Close()
					So(err, ShouldBeNil)
				})
			})
			Convey("It should return an error not found", func() {
				So(err, ShouldEqual, ErrProjectNotFound)
				Convey("The mock should complete successfully", func() {
					err = db.Close()
					So(err, ShouldBeNil)
				})
			})
		})

	})

}
