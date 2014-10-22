package service

import (
	"bufio"
	"bytes"
	"code.google.com/p/go.crypto/pbkdf2"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash"
	"io"
	"io/ioutil"
	"strconv"
	"time"
)

const (
	defaultKeySize = 32
)

var (
	ErrDecrypt = errors.New("error decrypting container")
	ErrMsgSize = errors.New("message too big")
)

const (
	// MaxMsgSize is the maximum size the (encoded) content of an Authorization container
	// can have
	MaxMsgSize = 4096
)

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
func (a *Authorization) ReadFrom(r io.Reader) (n int64, err error) {
	r = base64.NewDecoder(base64.StdEncoding, r)
	buf := bufio.NewReader(r)
	// timestamp
	timestamp, err := buf.ReadBytes('|')
	if err != nil {
		return
	}
	n += int64(len(timestamp))
	// remove delimiter
	timestamp = timestamp[:len(timestamp)-1]
	a.timestamp, err = strconv.ParseInt(string(timestamp), 10, 64)
	if err != nil {
		return
	}
	// read buffer parts
	var read int
	for _, binBuf := range [][]byte{a.salt, a.signature} {
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

// Expiry returns the expiry time
func (a *Authorization) Expiry() time.Time {
	return time.Unix(a.timestamp, 0)
}

// Expires sets the expiry time
func (a *Authorization) Expires(t time.Time) {
	a.timestamp = t.Unix()
}

// WriteTo writes the serialized authorization container to the given writer
//
// Prior to writing the container to a writer, it must be encoded using the Encode() method
func (a *Authorization) WriteTo(w io.Writer) (n int64, err error) {
	wr := base64.NewEncoder(base64.StdEncoding, w)
	defer wr.Close()
	ts := a.timestampBytes()
	written, err := wr.Write(append(ts, '|'))
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
	err = a.Sign(key)
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
