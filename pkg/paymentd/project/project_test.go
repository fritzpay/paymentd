package project_test

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	. "github.com/fritzpay/paymentd/pkg/paymentd/project"
	. "github.com/smartystreets/goconvey/convey"
)

func WithTestProject(db, prDB *sql.DB, f func(pr project.Project)) func() {
	return func() {
		princ := principal.Principal{}
		princ.Name = "payment_method_testprincipal"
		princ.CreatedBy = "test"
		err := principal.InsertPrincipalDB(prDB, &princ)
		So(err, ShouldBeNil)
		So(princ.ID, ShouldNotEqual, 0)
		So(princ.Empty(), ShouldBeFalse)

		Reset(func() {
			_, err = prDB.Exec("delete from principal where name = 'payment_method_testprincipal'")
			So(err, ShouldBeNil)
		})

		proj := project.Project{}
		proj.PrincipalID = princ.ID
		proj.Name = "payment_method_testproject"
		proj.CreatedBy = "test"
		err = project.InsertProjectDB(prDB, &proj)
		So(err, ShouldBeNil)
		So(proj.IsValid(), ShouldBeTrue)

		Reset(func() {
			_, err = prDB.Exec("delete from project where name = 'payment_method_testproject'")
			So(err, ShouldBeNil)
		})

		f(proj)
	}
}
