package config

import (
	"bytes"
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
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

func TestLoadSaveConfig(t *testing.T) {
	Convey("Given a default config", t, func() {
		cfg := DefaultConfig()

		Convey("When saving the config", func() {
			buf := bytes.NewBuffer(nil)
			err := WriteConfig(buf, cfg)

			Convey("It should complete successfully", func() {
				So(err, ShouldBeNil)
			})

			Convey("When loading the saved config", func() {
				loadedCfg, err := ReadConfig(buf)

				Convey("It should complete successfully", func() {
					So(err, ShouldBeNil)
				})

				Convey("The re-loaded config should match the saved config", func() {
					So(loadedCfg, ShouldHaveSameTypeAs, cfg)
					So(reflect.DeepEqual(loadedCfg, cfg), ShouldBeTrue)
				})
			})
		})
	})
}
