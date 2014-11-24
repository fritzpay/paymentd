// Authentication/authorization
package service

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"strconv"
	"sync"
	"time"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// MaxMsgSize is the maximum size the (encoded) content of an Authorization container
	// can have
	MaxMsgSize = 4096
)

const (
	// default capacity of keys in the keychain
	keychainLen = 16
	// default size of keys (generated) in bytes
	defaultKeySize = 32
)

var (
	ErrDecrypt       = errors.New("error decrypting container")
	ErrMsgSize       = errors.New("message too big")
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

// Authorization is a container which can hold arbitrary authorization data,
// can be encrypted and signed and safely passed between services. As long as those
// share the same keychain, they are able to access the encrypted data.
type Authorization struct {
	timestamp int64
	salt      []byte
	signature []byte
	rawMsg    []byte

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
func (a *Authorization) ReadFrom(r io.Reader) (int64, error) {
	var n int64
	var err error
	r = base64.NewDecoder(base64.StdEncoding, r)
	buf := bufio.NewReader(r)
	// timestamp
	timestamp, err := buf.ReadBytes('|')
	if err != nil {
		return n, err
	}
	n += int64(len(timestamp))
	// remove delimiter
	timestamp = timestamp[:len(timestamp)-1]
	a.timestamp, err = strconv.ParseInt(string(timestamp), 10, 64)
	if err != nil {
		return n, err
	}
	// read buffer parts
	var read int
	for _, binBuf := range [][]byte{a.salt, a.signature} {
		read, err = buf.Read(binBuf)
		if err != nil {
			return n, err
		}
		n += int64(read)
	}
	a.rawMsg, err = ioutil.ReadAll(buf)
	if err != nil {
		return n, err
	}
	n += int64(len(a.rawMsg))
	return n, err
}

// Expiry returns the expiry time
func (a *Authorization) Expiry() time.Time {
	if a.timestamp == 0 {
		return time.Time{}
	}
	return time.Unix(a.timestamp, 0)
}

// Expires sets the expiry time
func (a *Authorization) Expires(t time.Time) {
	a.timestamp = t.Unix()
}

// WriteTo writes the serialized authorization container to the given writer
//
// Prior to writing the container to a writer, it must be encoded using the Encode() method
func (a *Authorization) WriteTo(w io.Writer) (int64, error) {
	var n int64
	var err error
	wr := base64.NewEncoder(base64.StdEncoding, w)
	defer wr.Close()
	ts := a.timestampBytes()
	written, err := wr.Write(append(ts, '|'))
	if err != nil {
		return n, err
	}
	n += int64(written)
	for _, binBuf := range [][]byte{a.salt, a.signature, a.rawMsg} {
		written, err = wr.Write(binBuf)
		if err != nil {
			return n, err
		}
		n += int64(written)
	}
	return n, err
}

func (a *Authorization) Serialized() (string, error) {
	buf := bytes.NewBuffer(nil)
	_, err := a.WriteTo(buf)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (a *Authorization) timestampBytes() []byte {
	return []byte(strconv.FormatInt(a.timestamp, 10))
}

// Message implementing the Signable interface
func (a *Authorization) Message() ([]byte, error) {
	msg := append(a.timestampBytes(), a.salt...)
	msg = append(msg, a.rawMsg...)
	return msg, nil
}

// Signature implementing the Signed interface
func (a *Authorization) Signature() ([]byte, error) {
	return a.signature, nil
}

// HashFunc implementing the Signable interface
func (a *Authorization) HashFunc() func() hash.Hash {
	return a.H
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
	gzw := gzip.NewWriter(encoded)
	enc := json.NewEncoder(gzw)
	err := enc.Encode(a.Payload)
	gzw.Close()
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
	a.signature, err = Sign(a, key)
	return err
}

/// Decode decodes a container after it was read using the ReadFrom() method
func (a *Authorization) Decode(key []byte) error {
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
	gzr, err := gzip.NewReader(msgBuf)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(gzr)
	err = dec.Decode(&a.Payload)
	gzr.Close()
	return err
}
