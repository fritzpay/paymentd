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

func TestDatabaseConfig(t *testing.T) {
	Convey("Given an empty DatabaseConfig", t, func() {
		db := NewDatabaseConfig()

		Convey("When getting the database type of an empty DatabaseConfig", func() {
			Convey("It should panic", func() {
				So(func() { db.Type() }, ShouldPanic)
			})
		})

		Convey("When getting the DSN of an empty DatabaseConfig", func() {
			Convey("It should panic", func() {
				So(func() { db.DSN() }, ShouldPanic)
			})
		})

	})
}
