package config

import (
	"database/sql"
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/fritzpay/paymentd/pkg/testutil"
	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/crypto/bcrypt"
)

func TestConfigEntry(t *testing.T) {
	Convey("Given a DB connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})

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
				nonExistent, err := EntryByNameDB(db, "nonexistent")
				Convey("It should return an error", func() {
					So(err, ShouldNotBeNil)
					So(err, ShouldEqual, ErrEntryNotFound)
					So(nonExistent.Empty(), ShouldBeTrue)
				})
			})
		})
	}))
}

func TestConfigSetPassword(t *testing.T) {
	Convey("Given a payment DB connection", t, testutil.WithPaymentDB(t, func(db *sql.DB) {
		Reset(func() {
			db.Close()
		})

		Convey("Given a config with no password set", func() {
			_, err := db.Exec(fmt.Sprintf("delete from config where name = '%s'", ConfigNameSystemPassword))
			So(err, ShouldBeNil)

			Convey("When setting the password", func() {
				pw := SetPassword([]byte("password"))
				err := Set(db, pw)

				Reset(func() {
					_, err := db.Exec(fmt.Sprintf("delete from config where name = '%s'", ConfigNameSystemPassword))
					So(err, ShouldBeNil)
				})

				Convey("It should succeed", func() {
					So(err, ShouldBeNil)

					Convey("When retrieving the password entry", func() {
						val, err := EntryByNameDB(db, ConfigNameSystemPassword)
						So(err, ShouldBeNil)
						So(val.Empty(), ShouldBeFalse)

						Convey("It should match the password", func() {
							err := bcrypt.CompareHashAndPassword([]byte(val.Value), []byte("password"))
							So(err, ShouldBeNil)
						})
					})
				})
			})
		})
	}))
}
