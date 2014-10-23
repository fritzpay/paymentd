package metadata

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestMetadataEntry(t *testing.T) {
	Convey("Given an empty entry", t, func() {
		e := MetadataEntry{}

		Convey("When testing for emptiness", func() {
			empty := e.IsEmpty()
			Convey("It should be considered empty", func() {
				So(empty, ShouldBeTrue)
			})
		})
	})
}

func EachEntry(m Metadata, f func(e MetadataEntry)) func() {
	return func() {
		for _, e := range m {
			f(e)
		}
	}
}

func TestMetadataMapping(t *testing.T) {
	Convey("Given a key-value map", t, func() {
		m := map[string]string{
			"key":  "value",
			"test": "testval",
		}

		Convey("When creating metadata from the map", func() {
			meta := MetadataFromValues(m, "tester")

			Convey("It should create a correct metadata", func() {
				So(meta, ShouldNotBeNil)
				So(len(meta), ShouldEqual, len(m))
			})

			Convey("For each entry", func() {
				Convey("The creator should match", EachEntry(meta, func(e MetadataEntry) {
					So(e.CreatedBy, ShouldEqual, "tester")
				}))
			})

			Convey("When creating a map out of the metadata", func() {
				newMap := meta.Values()

				Convey("It should return a map", func() {
					So(newMap, ShouldNotBeNil)
				})
				Convey("It should match the original map", func() {
					So(reflect.DeepEqual(m, newMap), ShouldBeTrue)
				})
			})
		})
	})
}
