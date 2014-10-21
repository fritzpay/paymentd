package v1

import (
	. "github.com/smartystreets/goconvey/convey"
	"strings"
	"testing"
)

func TestInitPaymentRequest(t *testing.T) {
	Convey("Given a init payment request", t, func() {
		req := InitPaymentRequest{}

		Convey("When populated with test values", func() {
			req.ProjectKey = "abcdef123456"
			req.Ident = "testIdent"
			req.Amount.Int64 = 1234
			req.Subunits.Int8 = 2
			req.Currency = "EUR"
			req.Country = "DE"
			req.PaymentMethodID = 12
			req.Timestamp = 1234567
			req.Nonce = "testNonce"

			Convey("When creating a signature base string", func() {
				sig, err := req.SignatureBaseString()
				So(err, ShouldBeNil)

				Convey("It should match the expected signature", func() {
					expected := "abcdef123456testIdent12342EURDE121234567testNonce"
					So(sig, ShouldEqual, expected)
				})
			})

			Convey("When adding the optional locale value", func() {
				req.Locale = "en_US"

				Convey("When creating a signature base string", func() {
					sig, err := req.SignatureBaseString()
					So(err, ShouldBeNil)

					Convey("It should match the expected signature", func() {
						expected := "abcdef123456testIdent12342EURDE12en_US1234567testNonce"
						So(sig, ShouldEqual, expected)
					})
				})
			})
		})
	})
}

func TestInitPaymentRequestValidation(t *testing.T) {
	Convey("Given a init payment request", t, func() {
		req := InitPaymentRequest{}

		Convey("When populated with test values", func() {
			req.ProjectKey = "abcdef123456"
			req.Ident = "testIdent"
			req.Amount.Int64, req.Amount.Set = 1234, true
			req.Subunits.Int8, req.Subunits.Set = 2, true
			req.Currency = "EUR"
			req.Country = "DE"
			req.PaymentMethodID = 12
			req.Timestamp = 1234567
			req.Nonce = "testNonce"

			Convey("When validating without project key", func() {
				req.ProjectKey = ""
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "ProjectKey")
				})
			})

			Convey("When validating without ident", func() {
				req.Ident = ""
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Ident")
				})
			})

			Convey("When validating with a too large ident", func() {
				req.Ident = strings.Repeat("s", 200)
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Ident")
				})
			})

			Convey("When validating without an Amount", func() {
				req.Amount.Set = false
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Amount")
				})
			})

			Convey("When validating with a negative Amount", func() {
				req.Amount.Int64 = -1000
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Amount")
				})
			})

			Convey("When validating without a Subunit", func() {
				req.Subunits.Set = false
				err := req.Validate()
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldContainSubstring, "Subunits")
				})
			})
		})
	})
}
