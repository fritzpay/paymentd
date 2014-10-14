package service

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.crypto/pbkdf2"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"sync"
	"time"
)

const (
	keychainLen = 16
)

const (
	MaxMsgSize = 4096
)

var (
	ErrInvalidKey    = errors.New("invalid key")
	ErrNoKeys        = errors.New("no keys in keychain")
	ErrNoMatchingKey = errors.New("no encryption key for signature key")
	ErrBadContainer  = errors.New("bad container encoding")
	ErrDecrypt       = errors.New("error decrypting container")
	ErrMsgSize       = errors.New("message too big")
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

// Authorization is a container which can hold arbitrary authorization data,
// can be encrypted and signed and safely passed between services. As long as those
// share the same keychain, they are able to access the encrypted data.
type Authorization struct {
	timestamp int64
	salt      []byte
	signature []byte
	rawMsg    []byte

	Expiry  time.Time
	Payload map[string]interface{}
	H       func() hash.Hash
}

// NewAuthorization creates a new authorization container with the given Hash function
// for signing and key derivation
func NewAuthorization(h func() hash.Hash) *Authorization {
	return &Authorization{
		salt:      make([]byte, 8),
		signature: make([]byte, h().Size()),
		rawMsg:    make([]byte, 0, MaxMsgSize),

		Payload: make(map[string]interface{}),
		H:       h,
	}
}

// ReadFrom reads a serialized authorization container from the given reader
//
// After the container is read, it should be decoded using the Decode() method
func (a *Authorization) ReadFrom(r io.Reader) (n int64, err error) {
	r = base64.NewDecoder(base64.StdEncoding, r)
	buf := bufio.NewReader(r)
	// timestamp
	binBuf := make([]byte, binary.MaxVarintLen64)
	read, err := buf.Read(binBuf)
	if err != nil {
		return
	}
	n += int64(read)
	a.timestamp, read = binary.Varint(binBuf)
	if read <= 0 {
		err = ErrBadContainer
		return
	}
	// read buffer parts
	for _, binBuf = range [][]byte{a.salt, a.signature} {
		read, err = buf.Read(binBuf)
		if err != nil {
			return
		}
		n += int64(read)
	}
	a.rawMsg, err = ioutil.ReadAll(buf)
	if err != nil {
		return
	}
	n += int64(len(a.rawMsg))
	return
}

// WriteTo writes the serialized authorization container to the given writer
//
// Prior to writing the container to a writer, it must be encoded using the Encode() method
func (a *Authorization) WriteTo(w io.Writer) (n int64, err error) {
	wr := base64.NewEncoder(base64.StdEncoding, w)
	defer wr.Close()
	written, err := wr.Write(a.timestampBytes())
	if err != nil {
		return
	}
	n += int64(written)
	for _, binBuf := range [][]byte{a.salt, a.signature, a.rawMsg} {
		written, err = wr.Write(binBuf)
		if err != nil {
			return
		}
		n += int64(written)
	}
	return
}

func (a *Authorization) timestampBytes() []byte {
	a.timestamp = a.Expiry.Unix()
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, a.timestamp)
	return buf
}

// Message implementing the Signable interface
func (a *Authorization) Message() []byte {
	msg := append(a.timestampBytes(), a.salt...)
	msg = append(msg, a.rawMsg...)
	return msg
}

// Signature implementing the Signed interface
func (a *Authorization) Signature() []byte {
	return a.signature
}

// HashFunc implementing the Signable interface
func (a *Authorization) HashFunc() func() hash.Hash {
	return a.H
}

// Sign signs the container with the given key
func (a *Authorization) Sign(key []byte) error {
	mac := hmac.New(a.HashFunc(), key)
	_, err := mac.Write(a.Message())
	if err != nil {
		return err
	}
	a.signature = mac.Sum(nil)
	return nil
}

func (a *Authorization) generateSalt() error {
	_, err := rand.Read(a.salt)
	return err
}

func (a *Authorization) deriveKey(key []byte) []byte {
	const iter = 4096
	// AES-256
	const keyLen = 32
	return pbkdf2.Key(key, a.salt, iter, keyLen, a.HashFunc())
}

func (a *Authorization) encrypt(b cipher.Block, value []byte) ([]byte, error) {
	iv := make([]byte, b.BlockSize())
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(b, iv)
	stream.XORKeyStream(value, value)
	return append(iv, value...), nil
}

func (a *Authorization) decrypt(b cipher.Block, value []byte) ([]byte, error) {
	size := b.BlockSize()
	if len(value) <= size {
		return nil, ErrDecrypt
	}
	iv := value[:size]
	value = value[size:]
	stream := cipher.NewCTR(b, iv)
	stream.XORKeyStream(value, value)
	return value, nil
}

// Encode encodes the message, encrypting its contents and signing it with the given
// key
//
// Encode() must be called prior to writing it using the WriteTo() method, otherwise
// secret data might be written to the Writer
func (a *Authorization) Encode(key []byte) error {
	encoded := bytes.NewBuffer(nil)
	enc := json.NewEncoder(encoded)
	err := enc.Encode(a.Payload)
	if err != nil {
		return err
	}
	if encoded.Len() > MaxMsgSize {
		return ErrMsgSize
	}
	err = a.generateSalt()
	if err != nil {
		return err
	}
	blockKey := a.deriveKey(key)
	b, err := aes.NewCipher(blockKey)
	if err != nil {
		return err
	}
	a.rawMsg, err = a.encrypt(b, encoded.Bytes())
	if err != nil {
		return err
	}
	err = a.Sign(key)
	return err
}

/// Decode decodes a container after it was read using the ReadFrom() method
func (a *Authorization) Decode(key []byte) error {
	a.Expiry = time.Unix(a.timestamp, 0)

	blockKey := a.deriveKey(key)
	b, err := aes.NewCipher(blockKey)
	if err != nil {
		return err
	}
	a.rawMsg, err = a.decrypt(b, a.rawMsg)
	if err != nil {
		return err
	}
	msgBuf := bytes.NewReader(a.rawMsg)
	dec := json.NewDecoder(msgBuf)
	err = dec.Decode(&a.Payload)
	return err
}
