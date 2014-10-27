package payment

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// PaymentID represents an identifier for a payment
//
// It consists of a project ID and a payment ID
type PaymentID struct {
	ProjectID int64
	PaymentID int64
}

// ParsePaymentIDStr parses a given string of the format
//
//   123-12345
//
// into a PaymentID.
func ParsePaymentIDStr(str string) (PaymentID, error) {
	parts := strings.Split(str, "-")
	var id PaymentID
	var err error
	if len(parts) != 2 {
		return id, fmt.Errorf("invalid payment id str. expecting two parts")
	}
	id.ProjectID, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return id, fmt.Errorf("error parsing project id part: %v", err)
	}
	id.PaymentID, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return id, fmt.Errorf("error parsing payment id part: %v", err)
	}
	return id, nil
}

// String returns the string representation of a payment ID
//
// It is the inverse of the ParsePaymentIDStr
func (p PaymentID) String() string {
	return strconv.FormatInt(p.ProjectID, 10) + "-" + strconv.FormatInt(p.PaymentID, 10)
}

// MarshalJSON so it can be marshalled to JSON
func (p PaymentID) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

// UnmarshalJSON so it can be unmarshalled from JSON
func (p *PaymentID) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	*p, err = ParsePaymentIDStr(str)
	return err
}

type IDEncoder struct {
	max   *big.Int
	prime *big.Int
	inv   *big.Int
	xor   *big.Int
}

func NewIDEncoder(p, xor int64) (*IDEncoder, error) {
	m := &IDEncoder{}
	m.max = big.NewInt(math.MaxInt64)
	m.prime = big.NewInt(p)
	m.inv = new(big.Int)
	m.xor = big.NewInt(xor)
	g := new(big.Int)
	g.GCD(m.inv, nil, m.prime, new(big.Int).Add(m.max, big.NewInt(1)))
	if g.Int64() != 1 {
		return nil, errors.New("invalid p")
	}
	m.inv.Mod(m.inv, new(big.Int).Add(m.max, big.NewInt(1)))
	return m, nil
}

func (m *IDEncoder) Hide(i int64) int64 {
	hidden := big.NewInt(i)
	hidden.Mul(hidden, m.prime)
	hidden.And(hidden, m.max)
	hidden.Xor(hidden, m.xor)
	return hidden.Int64()
}

func (m *IDEncoder) Show(i int64) int64 {
	shown := big.NewInt(i)
	shown.Xor(shown, m.xor)
	shown.Mul(shown, m.inv)
	shown.And(shown, m.max)
	return shown.Int64()
}
