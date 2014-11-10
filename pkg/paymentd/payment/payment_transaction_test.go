package payment_test

import (
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestTransactionListBalance(t *testing.T) {
	Convey("Given a transaction list", t, func() {
		tl := payment.PaymentTransactionList([]*payment.PaymentTransaction{
			&payment.PaymentTransaction{
				Amount:   1000,
				Subunits: 2,
				Currency: "EUR",
			},
			&payment.PaymentTransaction{
				Amount:   1000,
				Subunits: 3,
				Currency: "EUR",
			},
		})

		Convey("When there is only one currency present", func() {
			Convey("When retrieving the balance", func() {
				b := tl.Balance()

				Convey("It should have one entry", func() {
					So(len(b), ShouldEqual, 1)
				})
				Convey("It should add the entries correctly", func() {
					So(b["EUR"].String(), ShouldEqual, "11.000")
					So(b["EUR"].IntegerPart(), ShouldEqual, "11")
					So(b["EUR"].DecimalPart(), ShouldEqual, "000")
				})
			})
		})

		Convey("When there are two currencies present", func() {
			tl = append(tl, &payment.PaymentTransaction{
				Amount:   1234,
				Subunits: 3,
				Currency: "USD",
			}, &payment.PaymentTransaction{
				Amount:   -1234,
				Subunits: 2,
				Currency: "USD",
			})

			Convey("When retrieving the balance", func() {
				b := tl.Balance()

				Convey("It should have two entries", func() {
					So(len(b), ShouldEqual, 2)
				})
				Convey("The sum should be correct", func() {
					So(b["USD"].String(), ShouldEqual, "-11.106")
				})
			})
		})
	})
}
