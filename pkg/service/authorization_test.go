package service

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	. "github.com/smartystreets/goconvey/convey"
	"hash"
	"reflect"
	"testing"
)

func TestKeychain(t *testing.T) {
	Convey("Given an empty keychain", t, func() {
		c := NewKeychain()

		Convey("When trying to get a hex key", func() {
			k, err := c.Key()
			Convey("It should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrNoKeys)
				So(k, ShouldBeBlank)
			})
		})

		Convey("When trying to get a binary key", func() {
			k, err := c.BinKey()
			Convey("It should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrNoKeys)
				So(k, ShouldBeNil)
			})
		})

		Convey("When adding a new hex key", func() {
			hexKey := "abcdef123456"
			err := c.AddKey(hexKey)
			Convey("It should complete successfully", func() {
				So(err, ShouldBeNil)
			})
			Convey("The keycount should be 1", func() {
				So(c.KeyCount(), ShouldEqual, 1)
			})

			Convey("When adding a second (binary) key", func() {
				bin := []byte{'a', 'c', 'f'}
				hexEnc := hex.EncodeToString(bin)
				c.AddBinKey(bin)
				Convey("The keycount should be 2", func() {
					So(c.KeyCount(), ShouldEqual, 2)
				})

				Convey("When retrieving a key from the keychain", func() {
					key, err := c.BinKey()
					Convey("It should return a key", func() {
						So(err, ShouldBeNil)
						So(key, ShouldNotBeEmpty)
					})
					Convey("The second key should have preference", func() {
						So(reflect.DeepEqual(key, bin), ShouldBeTrue)
					})
				})

				Convey("When retrieving the hex encoded binary key from the keychain", func() {
					key, err := c.Key()
					Convey("It should return a key", func() {
						So(err, ShouldBeNil)
						So(key, ShouldNotBeEmpty)
					})
					Convey("It should be correctly encoded", func() {
						So(key, ShouldEqual, hexEnc)
					})
				})
			})
		})

		Convey("When adding a badly encoded hex key", func() {
			badKey := "xfg"
			err := c.AddKey(badKey)

			Convey("It should return an error", func() {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, ErrInvalidKey)
			})
		})
	})
}

type TestMsg struct {
	msg []byte
	key []byte
}

func (t TestMsg) HashFunc() func() hash.Hash {
	return sha256.New
}

func (t TestMsg) Message() []byte {
	return t.msg
}

func (t TestMsg) Signature() []byte {
	mac := hmac.New(t.HashFunc(), t.key)
	_, err := mac.Write(t.Message())
	if err != nil {
		panic("error creating signature")
	}
	return mac.Sum(nil)
}

func TestKeychainCanMatchSigned(t *testing.T) {
	Convey("Given a keychain", t, func() {
		c := NewKeychain()

		Convey("Given a signed message", func() {
			msg := TestMsg{[]byte("test"), []byte("key")}

			Convey("When no keys are in the keychain", func() {
				So(c.KeyCount(), ShouldBeZeroValue)

				Convey("When retrieving a key for a signed message", func() {
					key, err := c.MatchKey(msg)
					Convey("It should return an error", func() {
						So(err, ShouldNotBeNil)
						So(err, ShouldEqual, ErrNoKeys)
						So(key, ShouldBeNil)
					})
				})
			})

			Convey("When keys are present in the keychain", func() {
				c.AddBinKey([]byte("one"))
				c.AddBinKey([]byte("two"))

				Convey("When the key is not in the keychain", func() {

					Convey("When retrieving a key for a signed message", func() {
						key, err := c.MatchKey(msg)
						Convey("It should return an error", func() {
							So(err, ShouldNotBeNil)
							So(key, ShouldBeNil)
							So(err, ShouldEqual, ErrNoMatchingKey)
						})
					})
				})

				Convey("When the key is in the keychain", func() {
					c.AddBinKey([]byte("key"))

					Convey("When retrieving a key for a signed message", func() {
						key, err := c.MatchKey(msg)
						Convey("It should return the matching key", func() {
							So(err, ShouldBeNil)
							So(reflect.DeepEqual(key, msg.key), ShouldBeTrue)
						})
					})
				})
			})
		})
	})
}

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
