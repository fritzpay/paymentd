package service

import (
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
