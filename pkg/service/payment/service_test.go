package payment_test

import (
	"database/sql"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/inconshreveable/log15.v2"
)

func WithService(ctx *service.Context, f func(s *paymentService.Service)) func() {
	return func() {
		s, err := paymentService.NewService(ctx)
		So(err, ShouldBeNil)

		f(s)
	}
}

func WithPayment(tx *sql.Tx, f func(p *payment.Payment)) func() {
	return func() {
		pr := &project.Project{
			ID:          1,
			PrincipalID: 1,
		}
		method, err := payment_method.PaymentMethodByIDTx(tx, 1)
		So(err, ShouldBeNil)
		So(method.Active(), ShouldBeTrue)
		So(method.ID, ShouldNotEqual, 0)

		err = payment_method.InsertPaymentMethodStatusTx(tx, method)
		So(err, ShouldBeNil)

		p := &payment.Payment{
			Created:  time.Now(),
			Ident:    "test001",
			Amount:   1234,
			Subunits: 2,
			Currency: "EUR",
		}
		err = p.SetProject(pr)
		So(err, ShouldBeNil)
		p.Config.SetCountry("DE")
		p.Config.SetLocale("en-US")
		p.Config.SetPaymentMethodID(method.ID)

		err = payment.InsertPaymentTx(tx, p)
		So(err, ShouldBeNil)
		err = payment.InsertPaymentConfigTx(tx, p)
		So(err, ShouldBeNil)

		p, err = payment.PaymentByIDTx(tx, p.PaymentID())
		So(err, ShouldBeNil)

		f(p)
	}
}

func TestPaymentNotification(t *testing.T) {
	Convey("Given a payment db connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Convey("Given a principal db connection", testutil.WithPrincipalDB(t, func(principalDB *sql.DB) {
			Convey("Given a transaction", func() {
				tx, err := db.Begin()
				So(err, ShouldBeNil)
				Reset(func() {
					err = tx.Rollback()
					So(err, ShouldBeNil)
				})

				Convey("Given a service context", testutil.WithContext(func(ctx *service.Context, logs <-chan *log15.Record) {
					ctx.SetPaymentDB(db, nil)
					ctx.SetPrincipalDB(principalDB, nil)

					Convey("Given a payment service", WithService(ctx, func(s *paymentService.Service) {

						Convey("Given a payment", WithPayment(tx, func(p *payment.Payment) {

							Convey("Given a test HTTP server", func() {
								srvOk := make(chan struct{})
								var req *http.Request
								var body []byte
								testSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
									req = r
									body, err = ioutil.ReadAll(r.Body)
									So(err, ShouldBeNil)
									close(srvOk)
								}))

								Reset(func() {
									testSrv.Close()
								})

								Convey("Given a test callback configuration", func() {
									testPk, err := project.ProjectKeyByKeyDB(principalDB, "testkey")
									So(err, ShouldBeNil)
									So(testPk.IsValid(), ShouldBeTrue)

									p.Config.SetCallbackURL(testSrv.URL)
									p.Config.SetCallbackAPIVersion("2")
									p.Config.SetCallbackProjectKey(testPk.Key)

									So(paymentService.CanCallback(&p.Config), ShouldBeTrue)

									Convey("When the payment has no transaction", func() {
										paymentTx, err := s.PaymentTransaction(tx, p)
										So(err, ShouldEqual, payment.ErrPaymentTransactionNotFound)
										So(paymentTx.Status.Valid(), ShouldBeFalse)

										Convey("When creating a transaction", func() {
											So(s.IsProcessablePayment(p), ShouldBeTrue)
											paymentTx = p.NewTransaction(payment.PaymentStatusOpen)
											err = s.SetPaymentTransaction(tx, paymentTx)

											Convey("It should succeed", func() {
												So(err, ShouldBeNil)
											})

											// Wrong implementation, notification should not occur inside a transaction
											//
											// 	Convey("A notification should be sent", func() {
											// 		select {
											// 		case <-srvOk:
											// 			So(req, ShouldNotBeNil)
											// 		case <-time.After(time.Second):
											// 			t.Errorf("request timeout on %s", testSrv.URL)
											// 			close(srvOk)
											// 		drain:
											// 			for {
											// 				select {
											// 				case msg := <-logs:
											// 					t.Logf("%v", msg)
											// 				default:
											// 					break drain
											// 				}
											// 			}
											// 		}

											// 		Convey("The notification should contain the transaction", func() {
											// 			not := &notification.Notification{}
											// 			dec := json.NewDecoder(bytes.NewBuffer(body))
											// 			err := dec.Decode(not)
											// 			So(err, ShouldBeNil)

											// 			So(not.TransactionTimestamp, ShouldNotEqual, 0)
											// 			So(not.Status, ShouldEqual, payment.PaymentStatusOpen)
											// 		})
											// 	})
										})
									})
								})
							})
						}))
					}))
				}))
			})
		}))
	}))
}
