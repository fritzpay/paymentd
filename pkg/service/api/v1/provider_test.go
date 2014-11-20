package v1

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/testutil"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"testing"
)

func TestGetProvider(t *testing.T) {
	Convey("Given a test context", t, testutil.WithContext(func(ctx *service.Context, logChan <-chan *log15.Record) {
		ctx.Config().API.ServeAdmin = true

		Convey("Given a service", WithService(ctx, logChan, func(s *Service, mx *mux.Router) {

			Convey("Given a request for the test provider", func() {
				req, err := http.NewRequest("GET", ServicePath+"/provider/fritzpay", nil)
				So(err, ShouldBeNil)

				rm := mux.RouteMatch{}
				match := mx.Match(req, &rm)
				So(match, ShouldBeTrue)

				Convey("Given a payment db", testutil.WithPaymentDB(t, func(db *sql.DB) {
					ctx.SetPaymentDB(db, nil)

					Convey("Given a valid authorization", WithAuthorization(mx, func(auth string) {
						req.Header.Set("Authorization", auth)

						Convey("When executing the request", func() {
							w := testutil.NewResponseWriter()
							mx.ServeHTTP(w, req)

							Convey("It should succeed", func() {
								So(w.HeaderWritten, ShouldBeTrue)
								So(w.StatusCode, ShouldEqual, http.StatusOK)
							})
						})
					}))
				}))
			})
		}))
	}))
}
