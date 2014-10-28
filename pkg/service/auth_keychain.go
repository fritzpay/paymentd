package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
)

const (
	keychainLen = 16
)

var (
	ErrInvalidKey    = errors.New("invalid key")
	ErrNoKeys        = errors.New("no keys in keychain")
	ErrNoMatchingKey = errors.New("no matching key for signature")
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
func (k *Keychain) KeyCount() int {
	k.m.RLock()
	c := len(k.keys)
	k.m.RUnlock()
	return c
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

// GenerateKey generates a random key, adds it to the keychain and returns the generated key
func (k *Keychain) GenerateKey() ([]byte, error) {
	key := make([]byte, defaultKeySize)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}
	k.pushKey(key)
	return key, nil
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
		if ok, err := IsAuthentic(s, key); err != nil {
			k.m.RUnlock()
			return nil, err
		} else if ok {
			k.m.RUnlock()
			return key, nil
		}
	}
	k.m.RUnlock()
	return nil, ErrNoMatchingKey
}
