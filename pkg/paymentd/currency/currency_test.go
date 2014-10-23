package currency

import (
	"encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestCurrencyJSONMarshalling(t *testing.T) {
	Convey("Given a currency", t, func() {
		curr := Currency{}
		curr.CodeISO4217 = "EUR"

		So(curr.IsEmpty(), ShouldBeFalse)

		Convey("When JSON-marshalling the currency", func() {
			j, err := json.Marshal(curr)
			So(err, ShouldBeNil)
			Convey("It should be encoded as a string literal", func() {
				So(string(j), ShouldEqual, "\"EUR\"")
			})
		})
	})
}
