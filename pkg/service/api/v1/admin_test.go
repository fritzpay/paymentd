package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"testing"
)

func WithAPI(ctx *service.Context, logChan <-chan *log15.Record, f func(a *AdminAPI)) func() {
	return func() {
		a := NewAdminAPI(ctx)

		testMsg := "testmsg"
		a.log.Info(testMsg)
		logMsg := <-logChan

		So(logMsg.Msg, ShouldEqual, testMsg)
		var pkg string
		for i := 0; i < len(logMsg.Ctx); i += 2 {
			if logMsg.Ctx[i].(string) != "pkg" {
				continue
			}
			pkg = logMsg.Ctx[i+1].(string)
			break
		}
		So(pkg, ShouldEqual, "github.com/fritzpay/paymentd/pkg/service/api/v1")

		f(a)
	}
}

func TestGetCredentialsWithBasicAuth(t *testing.T) {
	Convey("Given a new context", t, testutil.WithContext(func(ctx *service.Context, logChan <-chan *log15.Record) {
		Convey("Given a new API handler", WithAPI(ctx, logChan, func(a *AdminAPI) {

			Convey("Given a new get credentials request", func() {
				r, err := http.NewRequest("GET", "/user/credentials", nil)
				So(err, ShouldBeNil)

				Convey("When the request method is not GET", func() {
					r.Method = "POST"

					Convey("When the handler is called", func() {
						w := testutil.NewResponseWriter()
						a.GetAuthorization().ServeHTTP(w, r)
						Convey("The handler should respond with method not allowed", func() {
							So(w.HeaderWritten, ShouldBeTrue)
							So(w.StatusCode, ShouldEqual, http.StatusMethodNotAllowed)
						})
					})
				})

				Convey("When the authentication method is unknown", func() {
					r.URL.Path += "/unknown"

					Convey("When the handler is called", func() {
						w := testutil.NewResponseWriter()
						a.GetAuthorization().ServeHTTP(w, r)
						Convey("The handler should respond with a 404 (not found)", func() {
							So(w.HeaderWritten, ShouldBeTrue)
							So(w.StatusCode, ShouldEqual, http.StatusNotFound)
						})
					})
				})
			})
		}))
	}))
}
