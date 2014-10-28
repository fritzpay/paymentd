package payment

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

const (
	tokenBytes = 32
)

type PaymentToken struct {
	Token   string
	Created time.Time
	id      PaymentID
}

func NewPaymentToken(id PaymentID) (*PaymentToken, error) {
	if id.ProjectID == 0 {
		return nil, errors.New("payment id without project id")
	}
	if id.PaymentID == 0 {
		return nil, errors.New("payment id missing")
	}
	t := &PaymentToken{
		id:      id,
		Created: time.Now(),
	}
	err := t.GenerateToken()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (p *PaymentToken) GenerateToken() error {
	bin := make([]byte, tokenBytes)
	_, err := rand.Read(bin)
	if err != nil {
		return err
	}
	p.Token = hex.EncodeToString(bin)
	return nil
}

func (p *PaymentToken) Valid(timeout time.Duration) bool {
	now := time.Now()
	if now.Before(p.Created) {
		return false
	}
	if now.Sub(p.Created) > timeout {
		return false
	}
	return true
}
