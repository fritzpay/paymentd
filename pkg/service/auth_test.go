package service

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"reflect"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
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

		Convey("When generating a new key", func() {
			newKey, err := c.GenerateKey()
			Convey("It should return a new key", func() {
				So(err, ShouldBeNil)
				So(newKey, ShouldNotBeNil)
			})
			Convey("It should be in the keychain", func() {
				So(c.KeyCount(), ShouldEqual, 1)
			})

			Convey("When requesting a key from the keychain", func() {
				reqKey, err := c.BinKey()
				Convey("It should return a key", func() {
					So(err, ShouldBeNil)
					So(reqKey, ShouldNotBeNil)
				})
				Convey("It should match the generated key", func() {
					So(reflect.DeepEqual(newKey, reqKey), ShouldBeTrue)
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

func (t TestMsg) Message() ([]byte, error) {
	return t.msg, nil
}

func (t TestMsg) Signature() ([]byte, error) {
	mac := hmac.New(t.HashFunc(), t.key)
	msgBytes, err := t.Message()
	if err != nil {
		return nil, err
	}
	_, err = mac.Write(msgBytes)
	if err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
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
		key := []byte("testKey")
		auth := NewAuthorization(sha256.New)
		auth.Payload["test"] = "testValue"
		auth.Expires(time.Now())

		Convey("When encoding it", func() {
			err := auth.Encode(key)
			encryptedMsg := auth.rawMsg
			Convey("It should complete successfully", func() {
				So(err, ShouldBeNil)
				So(auth.rawMsg, ShouldNotBeNil)
			})

			Convey("When writing the encoded container", func() {
				buf := bytes.NewBuffer(nil)
				_, err = auth.WriteTo(buf)
				encoded := buf.Bytes()
				Convey("It should complete successfully", func() {
					So(err, ShouldBeNil)
					So(len(encoded), ShouldBeGreaterThan, 0)
					t.Logf("Encoded: %s", string(encoded))
				})

				Convey("Given a fresh authorization container", func() {
					newAuth := NewAuthorization(auth.HashFunc())

					Convey("When given the encoded container to read from", func() {
						buf = bytes.NewBuffer(encoded)
						_, err = newAuth.ReadFrom(buf)
						Convey("It should complete successfully", func() {
							So(err, ShouldBeNil)
						})
						Convey("The payload should be read", func() {
							So(newAuth.rawMsg, ShouldNotBeNil)
						})
						Convey("The encrypted message should be restored", func() {
							So(reflect.DeepEqual(newAuth.rawMsg, encryptedMsg), ShouldBeTrue)
						})

						Convey("When decoding the new container", func() {
							err = newAuth.Decode(key)
							Convey("It should complete successfully", func() {
								So(err, ShouldBeNil)
								So(newAuth.rawMsg, ShouldNotBeNil)
							})
							Convey("It should match the original authorization", func() {
								So(reflect.DeepEqual(newAuth.salt, auth.salt), ShouldBeTrue)
								So(newAuth.Expiry().Unix(), ShouldEqual, auth.Expiry().Unix())
								So(reflect.DeepEqual(newAuth.Payload, auth.Payload), ShouldBeTrue)
							})
						})
					})
				})
			})

		})
	})
}

func TestAuthorizationChain(t *testing.T) {
	Convey("Given a keychain with a set of keys", t, func() {
		chain := NewKeychain()
		chain.AddBinKey([]byte("one"))
		chain.AddBinKey([]byte("two"))
		chain.AddBinKey([]byte("three"))

		Convey("Given an authorization", func() {
			auth := NewAuthorization(sha256.New)
			auth.Payload["test"] = "testAuthChain"
			auth.Expires(time.Now())

			Convey("When retrieving a key from the keychain", func() {
				key, err := chain.BinKey()
				Convey("It should be successful", func() {
					So(err, ShouldBeNil)
					So(key, ShouldNotBeNil)
				})

				Convey("When using the key to encode the authorization", func() {
					err = auth.Encode(key)
					Convey("It should be successful", func() {
						So(err, ShouldBeNil)
					})

					Convey("When serializing the authorization", func() {
						buf := bytes.NewBuffer(nil)
						_, err = auth.WriteTo(buf)
						serialized := buf.Bytes()
						Convey("It should be successful", func() {
							So(err, ShouldBeNil)
						})

						Convey("Given a new authorization", func() {
							newAuth := NewAuthorization(auth.H)

							Convey("When reading the serialized original authorization", func() {
								_, err = newAuth.ReadFrom(bytes.NewReader(serialized))
								Convey("It should be successful", func() {
									So(err, ShouldBeNil)
								})

								Convey("When retrieving the key back from the keychain using the authorization", func() {
									newKey, err := chain.MatchKey(newAuth)
									Convey("It should be successful", func() {
										So(err, ShouldBeNil)
										So(newKey, ShouldNotBeNil)
									})

									Convey("When using the retrieved key to decode the authorization", func() {
										err = newAuth.Decode(newKey)
										Convey("It should be successful", func() {
											So(err, ShouldBeNil)
										})
										Convey("The authorization should match the original", func() {
											So(newAuth.Payload["test"], ShouldEqual, auth.Payload["test"])
										})
									})
								})
							})
						})
					})
				})
			})
		})
	})
}

func TestAuthorizationTime(t *testing.T) {
	Convey("Given an authorization", t, func() {
		auth := NewAuthorization(sha256.New)
		auth.Payload["test"] = "testValue"

		Convey("Given the authorization has no Expiry set", func() {
			So(auth.timestamp, ShouldEqual, 0)

			Convey("When retrieving the expiry", func() {
				exp := auth.Expiry()

				Convey("It should be treated as zero time", func() {
					So(exp.IsZero(), ShouldBeTrue)
				})
			})
		})
	})
}
