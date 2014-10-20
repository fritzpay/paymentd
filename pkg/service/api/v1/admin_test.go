package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"testing"
)

func WithService(ctx *service.Context, logChan <-chan *log15.Record, f func(s *Service, mux *http.ServeMux)) func() {
	return func() {
		ctx.Config().API.ServeAdmin = true

		So(ctx.Config().API.ServeAdmin, ShouldBeTrue)

		testMsg := "testmsg"
		ctx.Log().Info(testMsg)
		logMsg := <-logChan

		So(logMsg.Msg, ShouldEqual, testMsg)

		mux := http.NewServeMux()
		service := NewService(ctx, mux)

		f(service, mux)
	}
}

func TestGetCredentialsWithBasicAuth(t *testing.T) {
	Convey("Given a new context", t, testutil.WithContext(func(ctx *service.Context, logChan <-chan *log15.Record) {
		Convey("Given a new API service", WithService(ctx, logChan, func(s *Service, mux *http.ServeMux) {

			Convey("Given a new get credentials request", func() {
				r, err := http.NewRequest("GET", ServicePath+"/authorization", nil)
				So(err, ShouldBeNil)

				Convey("When the request method is PUT", func() {
					r.Method = "PUT"

					Convey("When the handler is called", func() {
						w := testutil.NewResponseWriter()
						mux.ServeHTTP(w, r)
						Convey("The handler should respond with method not allowed", func() {
							So(w.HeaderWritten, ShouldBeTrue)
							So(w.StatusCode, ShouldEqual, http.StatusMethodNotAllowed)
						})
					})
				})

				Convey("When the request method is DELETE", func() {
					r.Method = "DELETE"

					Convey("When the handler is called", func() {
						w := testutil.NewResponseWriter()
						mux.ServeHTTP(w, r)
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
						mux.ServeHTTP(w, r)
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
