package config

import (
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"math/rand"
	"strconv"
	"testing"
)

func TestConfigEntry(t *testing.T) {
	Convey("Given a DB connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Convey("Given a random entry name", func() {
			name := "test" + strconv.FormatInt(rand.Int63(), 10)

			Convey("When there is no entry with the given name", func() {
				_, err := db.Exec("delete from config where name = ?", name)
				So(err, ShouldBeNil)

				Convey("When entering an entry with a random value", func() {
					value := "entry" + strconv.FormatInt(rand.Int63(), 10)
					entry := Entry{
						Name:  name,
						Value: value,
					}
					err := InsertEntryDB(db, entry)

					Convey("It should succeed", func() {
						So(err, ShouldBeNil)
					})

					Convey("When retrieving the entry by the name", func() {
						entry, err := EntryByNameDB(db, name)

						Convey("It should succeed", func() {
							So(err, ShouldBeNil)
						})
						Convey("It should match the entered value", func() {
							So(entry.Value, ShouldEqual, value)
						})
					})
				})
			})
		})

		Convey("When there is no entry with a given name", func() {
			_, err := db.Exec("truncate config")
			So(err, ShouldBeNil)

			Convey("When selecting a nonexistent entry by name", func() {
				_, err := EntryByNameDB(db, "nonexistent")
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err, ShouldEqual, ErrEntryNotFound)
				})
			})
		})
	}))
}
