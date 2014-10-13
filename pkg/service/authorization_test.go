package service

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestEncodeDecodeAuthorization(t *testing.T) {
	Convey("Given an authorization", t, func() {
		Convey("When encoding it", func() {
			Convey("It should complete successfully", nil)
			Convey("When given it to decode", func() {
				Convey("It should complete successfully", nil)
				Convey("It should match the original authorization", nil)
			})
		})
	})
}
