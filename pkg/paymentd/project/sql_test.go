package project_test

import (
	"database/sql"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func WithTestProject(db, prDB *sql.DB, f func(pr *project.Project)) func() {
	return func() {
		princ := principal.Principal{}
		princ.Name = "project_testprincipal"
		princ.CreatedBy = "test"
		err := principal.InsertPrincipalDB(prDB, &princ)
		So(err, ShouldBeNil)
		So(princ.ID, ShouldNotEqual, 0)
		So(princ.Empty(), ShouldBeFalse)

		proj := &project.Project{}
		proj.PrincipalID = princ.ID
		proj.Name = "project_testproject"
		proj.CreatedBy = "test"
		err = project.InsertProjectDB(prDB, proj)
		So(err, ShouldBeNil)

		Reset(func() {
			_, err = prDB.Exec("delete from project where name = 'project_testproject'")
			So(err, ShouldBeNil)
			_, err = prDB.Exec("delete from principal where name = 'project_testprincipal'")
			So(err, ShouldBeNil)
		})

		f(proj)
	}
}

func TestProjectSQLMapping(t *testing.T) {
	Convey("Given a payment DB connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Convey("Given a principal DB connection", testutil.WithPrincipalDB(t, func(prDB *sql.DB) {
			Convey("Given a test project", WithTestProject(db, prDB, func(pr *project.Project) {

				Convey("When selecting the project without a present config", func() {
					selPr, err := project.ProjectByNameDB(prDB, pr.ID, pr.Name)
					So(err, ShouldBeNil)
					So(selPr.Empty(), ShouldBeFalse)
					Convey("The project config should not be set", func() {
						So(selPr.Config.IsSet(), ShouldBeFalse)
					})
				})

				Convey("Given a project config", func() {
					pr.Config.CallbackURL.String, pr.Config.CallbackURL.Valid = "http://www.example.com", true
					pr.Config.CallbackAPIVersion.String, pr.Config.CallbackAPIVersion.Valid = "1.2", true
					err := project.InsertProjectConfigDB(prDB, pr)
					So(err, ShouldBeNil)

					Reset(func() {
						_, err := prDB.Exec(fmt.Sprintf("delete from project_config where project_id = %d", pr.ID))
						So(err, ShouldBeNil)
					})

					Convey("When selecting the project", func() {
						selPr, err := project.ProjectByIdDB(prDB, pr.ID)
						So(err, ShouldBeNil)
						So(selPr.Empty(), ShouldBeFalse)

						Convey("The project config should be set", func() {
							So(selPr.Config.IsSet(), ShouldBeTrue)
						})
						Convey("The project config should match", func() {
							So(selPr.Config.CallbackURL.Valid, ShouldBeTrue)
							So(selPr.Config.CallbackURL.String, ShouldEqual, "http://www.example.com")
							So(selPr.Config.CallbackAPIVersion.Valid, ShouldBeTrue)
							So(selPr.Config.CallbackAPIVersion.String, ShouldEqual, "1.2")
						})
					})
				})
			}))
		}))
	}))
}
