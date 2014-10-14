package service

import (
	"crypto/hmac"
	"encoding/hex"
	"errors"
	"hash"
	"strconv"
	"sync"
	"time"
)

const (
	keychainLen = 16
)

var (
	ErrInvalidHashFunc   = errors.New("invalid hash func")
	ErrInvalidSignature  = errors.New("invalid signature")
	ErrSignatureKey      = errors.New("internal error while decoding signature key")
	ErrSignatureMismatch = errors.New("signature mismatch")
	ErrInvalidKey        = errors.New("invalid key")
	ErrNoKeys            = errors.New("no keys in keychain")
	ErrNoMatchingKey     = errors.New("no encryption key for signature key")
)

// Keychain stores (and rotates) keys for authorization container encryption
type Keychain struct {
	m    sync.RWMutex
	keys [][]byte
}

// NewKeychain creates an empty keychain
func NewKeychain() *Keychain {
	k := &Keychain{
		keys: make([][]byte, 0, keychainLen),
	}
	return k
}

// KeyCount returns the number of keys in the keychain
func (k *Keychain) KeyCount() (c int) {
	k.m.RLock()
	c = len(k.keys)
	k.m.RUnlock()
	return
}

func (k *Keychain) pushKey(key []byte) {
	k.m.Lock()
	keys := make([][]byte, 1, cap(k.keys)+keychainLen)
	keys[0] = key
	k.keys = append(keys, k.keys...)
	k.m.Unlock()
}

// AddKey adds a (hex-encoded) key to the keychain
func (k *Keychain) AddKey(newKey string) error {
	key, err := hex.DecodeString(newKey)
	if err != nil {
		return ErrInvalidKey
	}
	k.pushKey(key)
	return nil
}

// AddBinKey adds a binary key to the keychain
func (k *Keychain) AddBinKey(key []byte) {
	k.pushKey(key)
}

// Key returns a hex-encoded key from the keychain which can be used to encrypt authorization containers
func (k *Keychain) Key() (string, error) {
	key, err := k.BinKey()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(key), nil
}

// BinKey returns a binary key from the keychain which can be used to encrypt authorization containers
func (k *Keychain) BinKey() ([]byte, error) {
	k.m.RLock()
	if len(k.keys) == 0 {
		k.m.RUnlock()
		return nil, ErrNoKeys
	}
	key := k.keys[0]
	k.m.RUnlock()
	return key, nil
}

// MatchKey returns the key which was used to sign the Signed
// If no such key is in the keychain, it will return ErrNoMatchingKey
func (k *Keychain) MatchKey(s Signed) ([]byte, error) {
	k.m.RLock()
	if len(k.keys) == 0 {
		k.m.RUnlock()
		return nil, ErrNoKeys
	}
	for _, key := range k.keys {
		mac := hmac.New(s.HashFunc(), key)
		_, err := mac.Write(s.Message())
		if err != nil {
			k.m.RUnlock()
			return nil, err
		}
		if hmac.Equal(s.Signature(), mac.Sum(nil)) {
			k.m.RUnlock()
			return key, nil
		}
	}
	k.m.RUnlock()
	return nil, ErrNoMatchingKey
}

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

// Authorization is a container
type Authorization struct {
	timestamp int64
	rawMsg    []byte
	signature []byte

	Expiry  time.Time
	Payload interface{}
	H       func() hash.Hash
}

func (a *Authorization) Message() []byte {
	return append([]byte(strconv.FormatInt(a.timestamp, 10)), a.rawMsg...)
}

func (a *Authorization) Signature() []byte {
	return a.signature
}

func (a *Authorization) HashFunc() func() hash.Hash {
	return a.H
}
