package service

import (
	"database/sql"
	"testing"

	"github.com/fritzpay/paymentd/pkg/config"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
)

func WithContext(f func(ctx *Context)) func() {
	return func() {
		log := log15.New()
		cfg := config.DefaultConfig()
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
			db := &sql.DB{}
			ctx.SetPrincipalDB(db, nil)

			Convey("When no read-only connection is set", func() {

				Convey("When a read-only connection is requested", func() {
					reqDB := ctx.PrincipalDB(ReadOnly)

					Convey("It should return the write connection instead", func() {
						So(reqDB, ShouldNotBeNil)
						So(reqDB, ShouldEqual, db)
					})
				})
			})

			Convey("When both write and read-only connections are set", func() {
				roDB := &sql.DB{}
				ctx.SetPrincipalDB(db, roDB)

				Convey("When a write connection is requested", func() {
					reqDB := ctx.PrincipalDB()

					Convey("It should return the write connection", func() {
						So(reqDB, ShouldNotBeNil)
						So(reqDB, ShouldEqual, db)
					})
				})

				Convey("When a read-only connection is requested", func() {
					reqDB := ctx.PrincipalDB(ReadOnly)

					Convey("It should return the read-only connection", func() {
						So(reqDB, ShouldNotBeNil)
						So(reqDB, ShouldEqual, roDB)
					})
				})
			})
		})

		Convey("With set payment DB connections", func() {
			db := &sql.DB{}
			ctx.SetPaymentDB(db, nil)

			Convey("When no read-only connection is set", func() {

				Convey("When a read-only connection is requested", func() {
					reqDB := ctx.PaymentDB(ReadOnly)

					Convey("It should return the write connection instead", func() {
						So(reqDB, ShouldNotBeNil)
						So(reqDB, ShouldEqual, db)
					})
				})
			})

			Convey("When both write and read-only connections are set", func() {
				roDB := &sql.DB{}
				ctx.SetPaymentDB(db, roDB)

				Convey("When a write connection is requested", func() {
					reqDB := ctx.PaymentDB()

					Convey("It should return the write connection", func() {
						So(reqDB, ShouldNotBeNil)
						So(reqDB, ShouldEqual, db)
					})
				})

				Convey("When a read-only connection is requested", func() {
					reqDB := ctx.PaymentDB(ReadOnly)

					Convey("It should return the read-only connection", func() {
						So(reqDB, ShouldNotBeNil)
						So(reqDB, ShouldEqual, roDB)
					})
				})
			})
		})
	}))
}
