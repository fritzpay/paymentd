package json

import (
	j "encoding/json"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestUnmarshalInt64(t *testing.T) {
	tT := &struct {
		A RequiredInt64
		B int64 `json:",string"`
		C int64
	}{}
	jsonStr := []byte(`{}`)
	err := j.Unmarshal(jsonStr, tT)
	if err != nil {
		t.Error(err)
	}
	if tT.A.Set {
		t.Error("Expect A not to be set.")
	}
	if tT.B != 0 {
		t.Error("Expect B to be zero.")
	}
	if tT.C != 0 {
		t.Error("Expect C to be zero.")
	}
	jsonStr = []byte(`{"A":"12", "B":"23", "C":34}`)
	err = j.Unmarshal(jsonStr, tT)
	if err != nil {
		t.Error(err)
	}
	if !tT.A.Set {
		t.Error("Expect A to be set.")
	}
	if tT.A.Int64 != 12 {
		t.Errorf("Expect A to be %d, got %d", 12, tT.A.Int64)
	}
	if tT.B != 23 {
		t.Errorf("Expect B to be %d, got %d", 23, tT.B)
	}
	if tT.C != 34 {
		t.Errorf("Expect B to be %d, got %d", 34, tT.C)
	}
}

func TestRequiredInt64(t *testing.T) {
	Convey("Given a struct with a required int64", t, func() {
		ts := struct {
			A RequiredInt64
		}{}
		ts.A.Int64 = 1234

		Convey("When marshaling the struct", func() {
			marshalled, err := j.Marshal(ts)

			Convey("It should succeed", func() {
				So(err, ShouldBeNil)
			})
			Convey("It should be marshalled correctly", func() {
				expected := "{\"A\":\"1234\"}"
				So(string(marshalled), ShouldEqual, expected)
			})
		})
	})
}
