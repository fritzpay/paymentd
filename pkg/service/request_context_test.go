package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
	"time"
)

func TestRequestContextPurging(t *testing.T) {
	Convey("Given a new request", t, WithContext(func(ctx *Context) {
		r, err := http.NewRequest("GET", "www.example.com", nil)
		So(err, ShouldBeNil)

		Convey("Given a short timeout", func() {
			oldT := RequestContextTimeout
			RequestContextTimeout = func() time.Duration {
				return 100 * time.Millisecond
			}
			Reset(func() {
				RequestContextTimeout = oldT
			})

			Convey("When registering the context", func() {
				SetRequestContext(r, ctx)
				SetRequestContextVar(r, "test", "testvalue")
				Convey("It should be returned", func() {
					registeredCtx := RequestContext(r)
					So(registeredCtx, ShouldNotBeNil)
					testVal, ok := registeredCtx.Value("test").(string)
					So(ok, ShouldBeTrue)
					So(testVal, ShouldEqual, "testvalue")
				})

				Convey("When waiting for the timeout", func() {
					time.Sleep(110 * time.Millisecond)
					Convey("When retrieving the context", func() {
						registeredCtx := RequestContext(r)
						Convey("It should have been purged", func() {
							So(registeredCtx, ShouldBeNil)
						})
					})
				})
			})
		})

		Convey("When registering the context", func() {
			SetRequestContext(r, ctx)
			Convey("It should be returned", func() {
				registeredCtx := RequestContext(r)
				So(registeredCtx, ShouldNotBeNil)
			})

			Convey("When associating a var with the context", func() {
				SetRequestContextVar(r, "X", 10)

				Convey("When retrieving the context", func() {
					registeredCtx := RequestContext(r)
					Convey("It should match the registered context", func() {
						So(registeredCtx, ShouldNotBeNil)
						v, ok := registeredCtx.Value("X").(int)
						So(ok, ShouldBeTrue)
						So(v, ShouldEqual, 10)
					})
				})

				Convey("When clearing the context", func() {
					ClearRequestContext(r)

					Convey("When retrieving the context", func() {
						registeredCtx := RequestContext(r)
						Convey("It should not be registered", func() {
							So(registeredCtx, ShouldBeNil)
						})
					})
				})
			})
		})
	}))
}
