package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestContextSetup(t *testing.T) {
	Convey("Given a new ServiceContext", t, func() {

		Convey("When setting a principal DB with nil write connection", func() {

			Convey("It should panic", nil)

		})

		Convey("When setting a payment DB with nil write connection", func() {

			Convey("It shoud panic", nil)

		})

	})
}
