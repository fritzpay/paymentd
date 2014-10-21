package payment

import (
	"encoding/json"
	"fmt"
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
