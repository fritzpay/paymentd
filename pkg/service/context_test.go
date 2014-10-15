package service

import (
	"code.google.com/p/go.net/context"
	"github.com/fritzpay/paymentd/pkg/config"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/inconshreveable/log15.v2"
	"testing"
)

func WithContext(f func(ctx *Context)) func() {
	return func() {
		log := log15.New()
		cfg := config.Config{}
		ctx, err := NewContext(context.Background(), cfg, log)

		So(err, ShouldBeNil)

		f(ctx)
	}
}

func TestContextSetup(t *testing.T) {
	Convey("Given a new ServiceContext", t, WithContext(func(ctx *Context) {
		Convey("When setting a principal DB with nil write connection", func() {
			Convey("It should panic", func() {
				So(func() { ctx.SetPrincipalDB(nil, nil) }, ShouldPanic)
			})
		})

		Convey("When setting a payment DB with nil write connection", func() {
			Convey("It shoud panic", func() {
				So(func() { ctx.SetPaymentDB(nil, nil) }, ShouldPanic)
			})
		})
	}))
}

func TestDBReadOnlyHandling(t *testing.T) {
	Convey("Given a new service context", t, WithContext(func(ctx *Context) {

		Convey("With set principal DB connections", func() {

			Convey("When no read-only connection is set", func() {

				Convey("When a read-only connection is requested", func() {

					Convey("It should return the write connection instead", nil)

				})

			})

			Convey("When both write and read-only connections are set", func() {

				Convey("When a write connection is requested", func() {

					Convey("It should return the write connection", nil)

				})

				Convey("When a read-only connection is requested", func() {

					Convey("It should return the read-only connection", nil)

				})

			})

		})

		Convey("With set payment DB connections", func() {

			Convey("When no read-only connection is set", func() {

				Convey("When a read-only connection is requested", func() {

					Convey("It should return the write connection instead", nil)

				})

			})

			Convey("When both write and read-only connections are set", func() {

				Convey("When a write connection is requested", func() {

					Convey("It should return the write connection", nil)

				})

				Convey("When a read-only connection is requested", func() {

					Convey("It should return the read-only connection", nil)

				})

			})

		})

	}))

}
