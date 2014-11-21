package payment_test

import (
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/fritzpay/paymentd/pkg/testutil"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	. "github.com/smartystreets/goconvey/convey"
)

func WithTestProject(db, prDB *sql.DB, f func(pr *project.Project)) func() {
	return func() {
		princ, err := principal.PrincipalByNameDB(prDB, "testprincipal")
		So(err, ShouldBeNil)
		So(princ.ID, ShouldNotEqual, 0)
		So(princ.Empty(), ShouldBeFalse)

		proj, err := project.ProjectByPrincipalIDNameDB(prDB, princ.ID, "testproject")
		So(err, ShouldBeNil)

		f(proj)
	}
}

func WithTestPayment(tx *sql.Tx, pr *project.Project, f func(p *payment.Payment)) func() {
	return func() {
		p := &payment.Payment{}
		err := p.SetProject(pr)
		So(err, ShouldBeNil)

		p.Amount = 1234
		p.Subunits = 2
		p.Currency = "EUR"
		p.Created = time.Unix(1234, 0)

		err = payment.InsertPaymentTx(tx, p)
		So(err, ShouldBeNil)

		f(p)
	}
}

func TestPaymentAmountDecimal(t *testing.T) {
	Convey("Given a payment", t, func() {
		p := &payment.Payment{}
		p.Amount = 1234
		p.Subunits = 2

		Convey("When retrieving the decimal amount representation", func() {
			dec := p.Decimal()

			Convey("It should be correctly represented", func() {
				So(dec.String(), ShouldEqual, "12.34")
			})
		})
	})
}

func TestPaymentID(t *testing.T) {
	Convey("Given a payment ID string", t, func() {
		idStr := "1-1234"

		Convey("When parsing the payment ID", func() {
			id, err := payment.ParsePaymentIDStr(idStr)

			Convey("It should succeed", func() {
				So(err, ShouldBeNil)
				Convey("The parts should be parsed correctly", func() {
					So(id.PaymentID, ShouldEqual, 1234)
					So(id.ProjectID, ShouldEqual, 1)

					Convey("When getting a string representation", func() {
						str := id.String()
						Convey("It should match the original string", func() {
							So(str, ShouldEqual, idStr)
						})
					})
				})
			})
		})
	})
}

func TestIDEncoderModInv(t *testing.T) {
	Convey("Given a prime int64", t, func() {
		prime := int64(982450871)
		Convey("When calculating the modinv", func() {
			m, err := payment.NewIDEncoder(prime, rand.Int63())
			So(err, ShouldBeNil)
			Convey("Given a random int64", func() {
				r := rand.Int63()
				Convey("When encoding with the prime", func() {
					rn := m.Hide(r)
					t.Logf("%d encoded: %d", r, rn)
					Convey("When decoding given the inverse", func() {
						dec := m.Show(rn)
						Convey("It should match the original random int64", func() {
							So(dec, ShouldEqual, r)
						})
					})
				})
			})
		})
	})
}

func TestTransactionListBalance(t *testing.T) {
	Convey("Given a transaction list", t, func() {
		tl := payment.PaymentTransactionList([]*payment.PaymentTransaction{
			&payment.PaymentTransaction{
				Amount:   1000,
				Subunits: 2,
				Currency: "EUR",
			},
			&payment.PaymentTransaction{
				Amount:   1000,
				Subunits: 3,
				Currency: "EUR",
			},
		})

		Convey("When there is only one currency present", func() {
			Convey("When retrieving the balance", func() {
				b := tl.Balance()

				Convey("It should have one entry", func() {
					So(len(b), ShouldEqual, 1)
				})
				Convey("It should add the entries correctly", func() {
					So(b["EUR"].String(), ShouldEqual, "11.000")
					So(b["EUR"].IntegerPart(), ShouldEqual, "11")
					So(b["EUR"].DecimalPart(), ShouldEqual, "000")
				})
			})
		})

		Convey("When there are two currencies present", func() {
			tl = append(tl, &payment.PaymentTransaction{
				Amount:   1234,
				Subunits: 3,
				Currency: "USD",
			}, &payment.PaymentTransaction{
				Amount:   -1234,
				Subunits: 2,
				Currency: "USD",
			})

			Convey("When retrieving the balance", func() {
				b := tl.Balance()

				Convey("It should have two entries", func() {
					So(len(b), ShouldEqual, 2)
				})
				Convey("The sum should be correct", func() {
					So(b["USD"].String(), ShouldEqual, "-11.106")
				})
			})
		})
	})
}

func TestPaymentSQL(t *testing.T) {
	Convey("Given a payment DB", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})
		Convey("Given a principal DB", testutil.WithPrincipalDB(t, func(prDB *sql.DB) {
			Reset(func() {
				prDB.Close()
			})
			Convey("Given a test project", WithTestProject(db, prDB, func(proj *project.Project) {
				Convey("Given a transaction", func() {
					tx, err := db.Begin()
					So(err, ShouldBeNil)

					Reset(func() {
						err = tx.Rollback()
						So(err, ShouldBeNil)
					})

					Convey("Given a test payment", WithTestPayment(tx, proj, func(p *payment.Payment) {
						Convey("When selecting a payment by ident", func() {
							p2, err := payment.PaymentByProjectIDAndIdentTx(tx, proj.ID, p.Ident)
							Convey("It should succeed", func() {
								So(err, ShouldBeNil)
								Convey("It should match the original payment", func() {
									So(p2.ID(), ShouldEqual, p.ID())
								})
							})
						})

						Convey("Given the test payment has a transaction", func() {
							paymentTx := p.NewTransaction(payment.PaymentStatusPaid)
							paymentTx.Timestamp = time.Unix(9876, 0)
							err = payment.InsertPaymentTransactionTx(tx, paymentTx)
							So(err, ShouldBeNil)

							Convey("When selecting the payment", func() {
								p2, err := payment.PaymentByIDTx(tx, p.PaymentID())
								So(err, ShouldBeNil)

								Convey("The transaction values should be set in the payment", func() {
									So(p2.TransactionTimestamp.Unix(), ShouldEqual, 9876)
									So(p2.Status, ShouldEqual, payment.PaymentStatusPaid)
								})
							})
						})
					}))
				})
			}))
		}))
	}))
}

func TestPaymentTokenGenerationSQL(t *testing.T) {
	Convey("Given a payment DB", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})
		Convey("Given a principal DB", testutil.WithPrincipalDB(t, func(prDB *sql.DB) {
			Reset(func() {
				prDB.Close()
			})
			Convey("Given a test project", WithTestProject(db, prDB, func(proj *project.Project) {
				Convey("Given a transaction", func() {
					tx, err := db.Begin()
					So(err, ShouldBeNil)

					Reset(func() {
						err = tx.Rollback()
						So(err, ShouldBeNil)
					})

					Convey("Given a test payment", WithTestPayment(tx, proj, func(p *payment.Payment) {
						Convey("When generating a token for the payment", func() {
							t, err := payment.NewPaymentToken(p.PaymentID())
							So(err, ShouldBeNil)
							So(t.Valid(time.Minute), ShouldBeTrue)

							Convey("When inserting the token", func() {
								err = payment.InsertPaymentTokenTx(tx, t)

								Convey("It should succeed", func() {
									So(err, ShouldBeNil)

									Convey("Given a duplicate token", func() {
										t2 := *t
										So(t.Token, ShouldEqual, t2.Token)
										Convey("When inserting a duplicate token", func() {
											err = payment.InsertPaymentTokenTx(tx, &t2)
											Convey("It should succeed", func() {
												So(err, ShouldBeNil)
											})
											Convey("It should have regenerated the token", func() {
												So(t.Token, ShouldNotEqual, t2.Token)
											})
										})
									})
								})
							})
						})
					}))
				})
			}))
		}))
	}))
}

func TestPaymentMetadata(t *testing.T) {
	Convey("Given a payment DB", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})
		Convey("Given a principal DB", testutil.WithPrincipalDB(t, func(prDB *sql.DB) {
			Reset(func() {
				prDB.Close()
			})
			Convey("Given a test project", WithTestProject(db, prDB, func(proj *project.Project) {
				Convey("Given a transaction", func() {
					tx, err := db.Begin()
					So(err, ShouldBeNil)

					Reset(func() {
						err = tx.Rollback()
						So(err, ShouldBeNil)
					})

					Convey("Given a test payment", WithTestPayment(tx, proj, func(p *payment.Payment) {
						Convey("When adding metadata", func() {
							p.Metadata = map[string]string{"metadataEntry": "metadataValue"}
							err = payment.InsertPaymentMetadataTx(tx, p)
							So(err, ShouldBeNil)

							Convey("When retrieving the payment", func() {
								pRet, err := payment.PaymentByIDTx(tx, p.PaymentID())
								So(err, ShouldBeNil)

								Convey("It should return the metadata", func() {
									So(pRet.Metadata, ShouldNotBeNil)
									So(pRet.Metadata["metadataEntry"], ShouldEqual, "metadataValue")

									Convey("When adding an additional metadata entry", func() {
										pRet.Metadata["second"] = "me"
										err = payment.InsertPaymentMetadataTx(tx, pRet)
										So(err, ShouldBeNil)

										Convey("When retrieving the payment", func() {
											p, err = payment.PaymentByIDTx(tx, pRet.PaymentID())
											So(err, ShouldBeNil)

											Convey("It should have both entries", func() {
												So(p.Metadata, ShouldNotBeNil)
												So(len(p.Metadata), ShouldEqual, 2)
												So(p.Metadata["metadataEntry"], ShouldEqual, "metadataValue")
												So(p.Metadata["second"], ShouldEqual, "me")
											})
										})
									})
								})
							})
						})
					}))
				})
			}))
		}))
	}))
}
