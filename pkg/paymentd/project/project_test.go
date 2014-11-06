package project_test

import (
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestProjectConfig(t *testing.T) {
	Convey("Given a new project", t, func() {
		pr := &project.Project{}

		Convey("The config should initially have no values", func() {
			So(pr.Config.HasValues(), ShouldBeFalse)
		})

		Convey("When a config value is set", func() {
			pr.Config.SetProjectKey("testkey")

			Convey("The config should be considered to have values", func() {
				So(pr.Config.HasValues(), ShouldBeTrue)
			})
		})

		Convey("When marshalling the project config", func() {
			jsonStr, err := json.Marshal(pr.Config)
			Convey("It should be filled with null values", func() {
				So(err, ShouldBeNil)
				expect := `{"WebURL":null,"CallbackURL":null,"CallbackAPIVersion":null,"ProjectKey":null,"ReturnURL":null}`
				So(string(jsonStr), ShouldEqual, expect)
			})
		})

		Convey("Given a serialized JSON string", func() {
			cfgStr := `{"WebURL":"WebURL","CallbackURL":"CallbackURL","CallbackAPIVersion":"CallbackAPIVersion","ProjectKey":"ProjectKey","ReturnURL":"ReturnURL"}`

			Convey("When unmarshalling the JSON", func() {
				err := json.Unmarshal([]byte(cfgStr), &pr.Config)

				Convey("It should be correctly unmarshalled", func() {
					So(err, ShouldBeNil)
					So(pr.Config.HasValues(), ShouldBeTrue)
					So(pr.Config.WebURL.String, ShouldEqual, "WebURL")
				})

				Convey("When re-marshalling the config", func() {
					jsonStr, err := json.Marshal(pr.Config)

					Convey("It should match the original input", func() {
						So(err, ShouldBeNil)
						So(string(jsonStr), ShouldEqual, cfgStr)
					})
				})
			})
		})
	})
}
