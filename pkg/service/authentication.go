package service

import (
	"crypto/hmac"
	"hash"
)

// Signable is a type which can be signed
type Signable interface {
	Message() ([]byte, error)
	HashFunc() func() hash.Hash
}

// Signed is a type which has been signed
// The signature can be authenticated using the Signable interface and recreating the signature
type Signed interface {
	Signable
	Signature() ([]byte, error)
}

// IsAuthentic returns true if the signed message has a correct signature for the given key
func IsAuthentic(msg Signed, key []byte) (bool, error) {
	mac := hmac.New(msg.HashFunc(), key)
	msgBytes, err := msg.Message()
	if err != nil {
		return false, err
	}
	_, err = mac.Write(msgBytes)
	if err != nil {
		return false, err
	}
	sig, err := msg.Signature()
	if err != nil {
		return false, err
	}
	return hmac.Equal(sig, mac.Sum(nil)), nil
}

// Sign signs a signable message with the given key and returns the signature
func Sign(msg Signable, key []byte) ([]byte, error) {
	mac := hmac.New(msg.HashFunc(), key)
	msgBytes, err := msg.Message()
	if err != nil {
		return nil, err
	}
	_, err = mac.Write(msgBytes)
	if err != nil {
		return nil, err
	}
	return mac.Sum(nil), nil
}
