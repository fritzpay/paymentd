package web

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/fritzpay/paymentd/pkg/service/provider/fritzpay"
	"github.com/fritzpay/paymentd/pkg/testutil"
	"gopkg.in/inconshreveable/log15.v2"

	. "github.com/smartystreets/goconvey/convey"
)

func WithPayment(db *sql.DB, prDB *sql.DB, s *paymentService.Service, f func(p *payment.Payment, t *payment.PaymentToken)) func() {
	return func() {
		p := &payment.Payment{
			Created:  time.Now(),
			Ident:    "testPayment_" + fmt.Sprintf("%d", time.Now().UnixNano()),
			Amount:   1234,
			Subunits: 2,
			Currency: "EUR",
		}
		p.Config.SetCountry("DE")
		p.Config.SetPaymentMethodID(1)
		pr, err := project.ProjectByIDDB(prDB, 1)
		So(err, ShouldBeNil)
		err = p.SetProject(pr)
		So(err, ShouldBeNil)

		tx, err := db.Begin()
		So(err, ShouldBeNil)

		err = s.CreatePayment(tx, p)
		So(err, ShouldBeNil)
		err = s.SetPaymentConfig(tx, p)
		So(err, ShouldBeNil)
		t, err := s.CreatePaymentToken(tx, p)
		So(err, ShouldBeNil)

		err = tx.Commit()
		So(err, ShouldBeNil)

		f(p, t)
	}
}

func WithWebHandler(ctx *service.Context, f func(*Handler)) func() {
	return func() {
		h, err := NewHandler(ctx)
		So(err, ShouldBeNil)

		f(h)
	}
}

func TestPayment(t *testing.T) {
	Convey("Given a payment DB", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Convey("Given a principal DB", testutil.WithPrincipalDB(t, func(prDB *sql.DB) {
			Convey("Given a service context", testutil.WithContext(func(ctx *service.Context, logs <-chan *log15.Record) {
				ctx.SetPaymentDB(db, nil)
				ctx.SetPrincipalDB(prDB, nil)
				_, err := ctx.WebKeychain().GenerateKey()
				So(err, ShouldBeNil)
				ctx.Config().Web.URL = "file:///dev/null"

				Convey("Given a payment service", func() {
					s, err := paymentService.NewService(ctx)
					So(err, ShouldBeNil)

					Convey("With test web directories", func() {
						ctx.Config().Web.TemplateDir = os.TempDir()
						ctx.Config().Web.PubWWWDir = os.TempDir()
						ctx.Config().Provider.ProviderTemplateDir = os.TempDir()
						os.Mkdir(path.Join(os.TempDir(), fritzpay.FritzpayTemplateDir), os.FileMode(0644))

						Convey("Given a web handler", WithWebHandler(ctx, func(h *Handler) {
							Convey("Given an initialized payment", WithPayment(db, prDB, s, func(p *payment.Payment, token *payment.PaymentToken) {

								Convey("When the payment is already configured", func() {
									So(p.Config.Country.Valid, ShouldBeTrue)
									country := p.Config.Country.String

									Convey("Given not all configurations are set", func() {
										So(p.Config.Locale.Valid, ShouldBeFalse)

										Convey("Given the missing configuration can be determined", func() {
											req, err := http.NewRequest("GET", "http://example.com/payment", nil)
											So(err, ShouldBeNil)
											req.Header.Set("Accept-Language", "de-DE,de;q=0.8,en-US;q=0.6,en;q=0.4")

											Convey("Given a valid authorization cookie", func() {
												wr := testutil.NewResponseWriter()
												h.authenticatePaymentToken(wr, req, token.Token)
												So(wr.HeaderWritten, ShouldBeTrue)
												So(wr.StatusCode, ShouldEqual, http.StatusMovedPermanently)
												So(wr.Header().Get("Set-Cookie"), ShouldNotBeEmpty)
												cookieP := strings.Split(wr.Header().Get("Set-Cookie"), ";")
												req.Header.Set("Cookie", cookieP[0])

												Convey("When opening the payment", func() {
													wr = testutil.NewResponseWriter()
													h.ServeHTTP(wr, req)

													Convey("It should succeed", func() {
														So(wr.StatusCode, ShouldEqual, http.StatusOK)

														Convey("When retrieving the payment", func() {
															p, err := payment.PaymentByIDDB(db, p.PaymentID())
															So(err, ShouldBeNil)

															Convey("The configuration should be extended with the missing configuration", func() {
																So(p.Config.Locale.Valid, ShouldBeTrue)
															})

															Convey("The original configuration should not be touched", func() {
																So(p.Config.Country.String, ShouldEqual, country)
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
					})
				})
			}))
		}))
	}))
}
