package web

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPayment(t *testing.T) {
	Convey("Given an initialized payment", t, func() {

		Convey("When the payment is already configured", func() {

			Convey("Given not all configurations are set", func() {

				Convey("Given the missing configuration can be determined", func() {

					Convey("When opening the payment", func() {

						Convey("The configuration should be extended with the missing configuration", nil)

						Convey("The original configuration should not be touched", nil)

					})

				})

			})

		})

	})

}
