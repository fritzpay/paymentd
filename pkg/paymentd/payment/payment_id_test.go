package payment

import (
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
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

func TestIDEncoderModInv(t *testing.T) {
	Convey("Given a prime int64", t, func() {
		prime := int64(982450871)
		Convey("When calculating the modinv", func() {
			m, err := NewIDEncoder(prime, rand.Int63())
			So(err, ShouldBeNil)
			t.Logf("prime %d inv %d", prime, m.inv.Int64())
			Convey("Given a random int64", func() {
				r := rand.Int63()
				Convey("When encoding with the prime", func() {
					rn := m.Hide(r)
					t.Logf("%d encoded: %d", r, rn)
					Convey("When decoding given the inverse", func() {
						dec := m.Show(rn)
						Convey("It should match the original random int64", func() {
							So(dec, ShouldEqual, r)
						})
					})
				})
			})
		})
	})

}
