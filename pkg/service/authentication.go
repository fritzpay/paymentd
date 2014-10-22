package service

import (
	"crypto/hmac"
	"hash"
)

// Signable is a type which can be signed
type Signable interface {
	Message() []byte
	HashFunc() func() hash.Hash
}

// Signed is a type which has been signed
// The signature can be authenticated using the Signable interface and recreating the signature
type Signed interface {
	Signable
	Signature() []byte
}

// IsAuthentic returns true if the signed message has a correct signature for the given key
func IsAuthentic(msg Signed, key []byte) (bool, error) {
	mac := hmac.New(msg.HashFunc(), key)
	_, err := mac.Write(msg.Message())
	if err != nil {
		return false, err
	}
	return hmac.Equal(msg.Signature(), mac.Sum(nil)), nil
}
