package env

import (
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/inconshreveable/log15.v2"
	golog "log"
	"testing"
)

type testHandler struct {
	record *log15.Record
}

func (t *testHandler) Log(r *log15.Record) error {
	t.record = r
	return nil
}

func TestLogPkgIsBridged(t *testing.T) {
	handler := &testHandler{}

	Convey("Given a new environment", t, func() {
		InitLog()
		Log.SetHandler(handler)

		Convey("When a log message is created using the go log pkg", func() {
			msg := "Log message"
			golog.Print(msg)

			Convey("The record should be received by the log15 handler", func() {
				So(handler.record, ShouldNotBeNil)
			})

			Convey("The record lvl should be Info", func() {
				So(handler.record.Lvl, ShouldEqual, log15.LvlInfo)
			})

			Convey("The actual message should be present", func() {
				var actual string
				for i, c := range handler.record.Ctx {
					if key, ok := c.(string); !ok {
						continue
					} else if key == "message" {
						actual = handler.record.Ctx[i+1].(string)
					}
				}
				So(actual, ShouldNotBeBlank)

				Convey("The message should be passed", func() {
					So(actual, ShouldContainSubstring, msg)
				})
			})
		})
	})
}
