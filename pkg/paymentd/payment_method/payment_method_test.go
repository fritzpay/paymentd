package payment_method

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPaymentMethodStatus(t *testing.T) {
	Convey("Given an empty payment method status", t, func() {
		st := methodStatus("")

		Convey("When retrieving the string value of the status", func() {
			str := st.String()

			Convey("It should be \"invalid\"", func() {
				So(str, ShouldEqual, "invalid")
			})
		})
	})

	Convey("Given an active status", t, func() {
		st := PaymentMethodStatusActive

		Convey("When retrieving the string value of the status", func() {
			str := st.String()

			Convey("It should be \"active\"", func() {
				So(str, ShouldEqual, "active")
			})
		})
	})
}
