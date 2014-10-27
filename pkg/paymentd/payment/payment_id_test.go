package payment

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestPaymentID(t *testing.T) {
	Convey("Given a payment ID string", t, func() {
		idStr := "1-1234"

		Convey("When parsing the payment ID", func() {
			id, err := ParsePaymentIDStr(idStr)

			Convey("It should succeed", func() {
				So(err, ShouldBeNil)
				Convey("The parts should be parsed correctly", func() {
					So(id.PaymentID, ShouldEqual, 1234)
					So(id.ProjectID, ShouldEqual, 1)

					Convey("When getting a string representation", func() {
						str := id.String()
						Convey("It should match the original string", func() {
							So(str, ShouldEqual, idStr)
						})
					})
				})
			})
		})
	})
}
