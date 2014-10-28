package v1

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/testutil"
	"github.com/gorilla/mux"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"strings"
	"testing"
)

func WithService(ctx *service.Context, logChan <-chan *log15.Record, f func(s *Service, mux *mux.Router)) func() {
	return func() {
		testMsg := "testmsg"
		ctx.Log().Info(testMsg)
		logMsg := <-logChan

		So(logMsg.Msg, ShouldEqual, testMsg)

		mux := mux.NewRouter()
		service, err := NewService(ctx, mux)
		So(err, ShouldBeNil)

		f(service, mux)
	}
}

func TestServiceSetup(t *testing.T) {
	Convey("Given a new context", t, testutil.WithContext(func(ctx *service.Context, logChan <-chan *log15.Record) {

		Convey("When the admin API is active", func() {
			ctx.Config().API.ServeAdmin = true

			Convey("Given a new service", WithService(ctx, logChan, func(s *Service, mx *mux.Router) {

				Convey("The admin API routes should be registered", func() {
					r, err := http.NewRequest("GET", ServicePath+"/authorization", nil)
					So(err, ShouldBeNil)

					rm := mux.RouteMatch{}
					match := mx.Match(r, &rm)

					So(match, ShouldBeTrue)
				})
			}))
		})

		Convey("When the config does not request the admin API to be active", func() {
			ctx.Config().API.ServeAdmin = false

			So(ctx.Config().API.ServeAdmin, ShouldBeFalse)

			Convey("Given a new service", WithService(ctx, logChan, func(s *Service, mx *mux.Router) {

				Convey("The admin registered log message should not be present", func() {
					var logMessagePresent bool
				drain:
					for {
						select {
						case msg := <-logChan:
							if strings.Contains(msg.Msg, "admin API") {
								logMessagePresent = true
							}
						default:
							So(logMessagePresent, ShouldBeFalse)
							break drain
						}
					}
				})

				Convey("Then the admin API routes should not be registered", func() {
					r, err := http.NewRequest("GET", ServicePath+"/authorization", nil)
					So(err, ShouldBeNil)

					rm := mux.RouteMatch{}
					match := mx.Match(r, &rm)

					So(match, ShouldBeFalse)
				})
			}))
		})
	}))
}
