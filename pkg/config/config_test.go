package config

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestReadConfig(t *testing.T) {
	Convey("Given a config", t, func() {
		buf := bytes.NewBuffer(nil)

		Convey("When the config Reader content is erroneous", func() {
			buf.WriteString("feeffefefe")
			_, err := ReadConfig(buf)

			Convey("The ReadConfig method should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}
