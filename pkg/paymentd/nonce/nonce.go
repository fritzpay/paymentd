package nonce

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

const (
	NonceBytes = 32
)

type Nonce struct {
	Nonce   string
	Created time.Time
}

func New() (*Nonce, error) {
	n := &Nonce{}
	err := n.Generate()
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (n *Nonce) Generate() error {
	b := make([]byte, NonceBytes)
	_, err := rand.Read(b)
	if err != nil {
		return err
	}
	str := base64.StdEncoding.EncodeToString(b)
	n.Nonce = string(str[:NonceBytes])
	return nil
}
