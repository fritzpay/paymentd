package payment_method

import (
	"database/sql"
	"testing"

	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/paymentd/provider"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
)

func TestPaymentMethodSQL(t *testing.T) {
	Convey("Given a payment DB connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})
		Convey("Given a principal DB connection", testutil.WithPrincipalDB(t, func(prDB *sql.DB) {
			Reset(func() {
				prDB.Close()
			})
			Convey("Given a test principal", func() {
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

				Convey("Given a test project", func() {
					proj := project.Project{}
					proj.PrincipalID = princ.ID
					proj.Name = "payment_method_testproject"
					proj.CreatedBy = "test"
					err := project.InsertProjectDB(prDB, &proj)
					So(err, ShouldBeNil)

					Reset(func() {
						_, err = prDB.Exec("delete from project where name = 'payment_method_testproject'")
						So(err, ShouldBeNil)
					})

					Convey("Given a transaction", func() {
						tx, err := db.Begin()
						So(err, ShouldBeNil)

						Reset(func() {
							err = tx.Rollback()
							So(err, ShouldBeNil)
						})

						Convey("Given a test provider exists", func() {
							pr, err := provider.ProviderByIDTx(tx, 1)
							So(err, ShouldBeNil)
							So(pr.ID, ShouldEqual, 1)

							Convey("When retrieving a nonexistent payment method", func() {
								_, err = PaymentMethodByProjectIDProviderIDMethodKey(db, proj.ID, pr.ID, "test")
								Convey("It should return a not found error", func() {
									So(err, ShouldEqual, ErrPaymentMethodNotFound)
								})
							})

							Convey("When inserting a new payment method", func() {
								pm := &Method{}
								pm.ProjectID = proj.ID
								pm.Provider.ID = pr.ID
								pm.MethodKey = "test"
								pm.CreatedBy = "test"

								err = InsertPaymentMethodTx(tx, pm)
								So(err, ShouldBeNil)
								So(pm.ID, ShouldNotEqual, 0)

								Convey("When setting the status to active", func() {
									pm.Status = PaymentMethodStatusActive
									pm.CreatedBy = "test"

									err = InsertPaymentMethodStatusTx(tx, pm)
									So(err, ShouldBeNil)

									Convey("When retrieving the payment method", func() {
										newPm, err := PaymentMethodByIDTx(tx, pm.ID)
										So(err, ShouldBeNil)

										Convey("The retrieved payment method should match", func() {
											So(newPm.Status, ShouldEqual, PaymentMethodStatusActive)
										})
									})
								})

								Convey("When setting metadata", func() {
									pm.Metadata = map[string]string{
										"name": "value",
										"test": "check",
									}
									err = InsertPaymentMethodMetadataTx(tx, pm, "metatest")
									So(err, ShouldBeNil)

									Convey("When selecting metadata", func() {
										metadata, err := PaymentMethodMetadataTx(tx, pm)
										So(err, ShouldBeNil)

										Convey("It should match", func() {
											So(metadata, ShouldNotBeNil)
											So(metadata["name"], ShouldEqual, "value")
											So(metadata["test"], ShouldEqual, "check")
										})
									})
								})
							})
						})
					})
				})
			})
		}))
	}))
}
