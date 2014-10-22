package env

import (
	"errors"
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
	Convey("Given a new environment", t, func() {
		handler := &testHandler{}
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
				So(actual, ShouldContainSubstring, msg)
			})
		})
	})
}

type TestStringer string

func (t TestStringer) String() string {
	return string(t)
}

func TestDaemonLogFmt(t *testing.T) {
	Convey("Given a handler with the DaemonLog format", t, func() {
		handler := &testHandler{}
		Log.SetHandler(handler)

		Convey("Given a log message with a log level", func() {
			Convey("When logging a log level Crit", func() {
				msg := "crit message"
				Log.Crit(msg)

				Convey("The log message should be prefixed with SD_CRIT", func() {
					logStr := string(DaemonFormat().Format(handler.record))
					So(logStr, ShouldStartWith, sdCrit)
				})
			})
			Convey("When logging a log level Error", func() {
				msg := "error message"
				Log.Error(msg)

				Convey("The log message should be prefixed with SD_ERR", func() {
					logStr := string(DaemonFormat().Format(handler.record))
					So(logStr, ShouldStartWith, sdErr)
				})
			})
			Convey("When logging a log level Warn", func() {
				msg := "warn message"
				Log.Warn(msg)

				Convey("The log message should be prefixed with SD_WARNING", func() {
					logStr := string(DaemonFormat().Format(handler.record))
					So(logStr, ShouldStartWith, sdWarning)
				})
			})
			Convey("When logging a log level Info", func() {
				msg := "info message"
				Log.Info(msg)

				Convey("The log message should be prefixed with SD_INFO", func() {
					logStr := string(DaemonFormat().Format(handler.record))
					So(logStr, ShouldStartWith, sdInfo)
				})
			})
			Convey("When logging a log level Debug", func() {
				msg := "debug message"
				Log.Debug(msg)

				Convey("The log message should be prefixed with SD_DEBUG", func() {
					logStr := string(DaemonFormat().Format(handler.record))
					So(logStr, ShouldStartWith, sdDebug)
				})
			})
		})

		Convey("When logging a complex string", func() {
			str := "this\\should\tbe\r\nescaped\""
			Log.Debug("escape", log15.Ctx{"this": str})

			Convey("The log message should be properly escaped", func() {
				expect := "this=\"this\\\\should\\tbe\\r\\nescaped\\\"\""
				logStr := string(DaemonFormat().Format(handler.record))
				So(logStr, ShouldContainSubstring, expect)
			})
		})

		Convey("When logging a complex Stringer", func() {
			str := TestStringer("this\\should\tbe\r\nescaped\"")
			Log.Debug("escape", log15.Ctx{"this": str})

			Convey("The log message should be properly escaped", func() {
				expect := "this=\"this\\\\should\\tbe\\r\\nescaped\\\"\""
				logStr := string(DaemonFormat().Format(handler.record))
				So(logStr, ShouldContainSubstring, expect)
			})
		})

		Convey("When logging a complex error", func() {
			err := errors.New("this\\should\tbe\r\nescaped\"")
			Log.Debug("escape", log15.Ctx{"err": err})

			Convey("The log message should be properly escaped", func() {
				expect := "err=\"this\\\\should\\tbe\\r\\nescaped\\\"\""
				logStr := string(DaemonFormat().Format(handler.record))
				So(logStr, ShouldContainSubstring, expect)
			})
		})
	})
}
