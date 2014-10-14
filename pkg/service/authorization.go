package service

import (
	"bufio"
	"compress/gzip"
	"crypto/aes"
	"crypto/hmac"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"strconv"
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
	ErrNoKeys            = errors.New("no keys in keychain")
	ErrNoEncryptionKey   = errors.New("no encryption key for signature key")
)

type Keychain struct {
	keys     map[string][]byte
	keyOrder []string
}

func NewKeychain() *Keychain {
	k := &Keychain{
		keys:     make(map[string][]byte),
		keyOrder: make([]string, 0, keychainLen),
	}
	return k
}

func (k *Keychain) AddPair(sigKey, blockKey []byte) error {
	sigKeyStr := base64.StdEncoding.EncodeToString(sigKey)
	if _, ok := k.keys[sigKeyStr]; ok {
		return errors.New("signature key is already assigned to a block key")
	}
	// test if blockKey is a valid AES block key
	_, err := aes.NewCipher(blockKey)
	if err != nil {
		return err
	}
	k.keyOrder = append(k.keyOrder, sigKeyStr)
	k.keys[sigKeyStr] = blockKey
	return nil
}

func (k *Keychain) EncryptionKey(s Signed) ([]byte, error) {
	if s.HashFunc() == nil {
		return nil, ErrInvalidHashFunc
	}
	if len(k.keyOrder) == 0 {
		return nil, ErrNoKeys
	}
	messageSig, err := base64.StdEncoding.DecodeString(s.Signature())
	if err != nil {
		return nil, ErrInvalidSignature
	}
	for i := len(k.keyOrder) - 1; i >= 0; i-- {
		sigKey, err := base64.StdEncoding.DecodeString(k.keyOrder[i])
		if err != nil {
			return nil, ErrSignatureKey
		}
		mac := hmac.New(s.HashFunc(), sigKey)
		_, err = mac.Write(s.Message())
		if err != nil {
			return nil, err
		}
		if hmac.Equal(messageSig, mac.Sum(nil)) {
			if blockKey, ok := k.keys[s.Signature()]; !ok {
				return nil, ErrNoEncryptionKey
			} else {
				return blockKey, nil
			}
		}
	}
	return nil, ErrSignatureMismatch
}

func (k *Keychain) Pair() (string, []byte, error) {
	if len(k.keyOrder) == 0 {
		return "", nil, ErrNoKeys
	}
	sigKey := k.keyOrder[len(k.keyOrder)-1]
	if key, ok := k.keys[sigKey]; !ok {
		return "", nil, ErrNoEncryptionKey
	} else {
		return sigKey, key, nil
	}
}

type Signable interface {
	Message() []byte
	HashFunc() func() hash.Hash
}

type Signed interface {
	Signable
	Signature() string
}

type Authorization struct {
	Expiry    time.Time
	expiry    []byte
	rawMsg    []byte
	UserID    string
	Payload   interface{}
	signature []byte
	hashFunc  func() hash.Hash
}

func (a *Authorization) WriteAuthorization(w io.Writer) error {
	w = base64.NewEncoder(base64.StdEncoding, w)
	w = gzip.NewWriter(w)
	fmt.Fprintf(w, "%d|", a.Expiry.Unix())
	gobEnc := gob.NewEncoder(w)
	err := gobEnc.Encode(a.Payload)
	if err != nil {
		return err
	}
	_, err = w.Write('|')
	if err != nil {
		return err
	}
	return nil
}

func ReadAuthorization(r io.Reader) (*Authorization, error) {
	var err error
	a := &Authorization{}
	gzr, err := gzip.NewReader(base64.NewDecoder(base64.StdEncoding, r))
	if err != nil {
		return nil, err
	}
	buf := bufio.NewReader(gzr)

	a.expiry, err = buf.ReadBytes('|')
	if err != nil {
		return nil, err
	}
	// remove delimiter
	a.expiry = a.expiry[:len(a.expiry)-1]

	timestamp, err := strconv.ParseInt(string(a.expiry), 10, 64)
	if err != nil {
		return nil, err
	}
	a.Expiry = time.Unix(timestamp, 0)

	a.rawMsg, err = buf.ReadBytes('|')
	if err != nil {
		return nil, err
	}
	a.rawMsg = a.rawMsg[:len(a.rawMsg)-1]

	a.signature, err = ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func (a *Authorization) Message() []byte {
	return append(a.expiry, a.rawMsg...)
}

func (a *Authorization) Signature() []byte {
	return a.signature
}

func (a *Authorization) HashFunc() func() hash.Hash {
	return a.hashFunc
}
