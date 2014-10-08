package v1

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestServiceSetup(t *testing.T) {
	Convey("Given a new service", t, func() {

		Convey("When the config does not request the admin API to be active", func() {

			Convey("Then the admin API routes should not be registered", nil)

		})

	})
}
