package payment

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

type PaymentID struct {
	ProjectID int64
	PaymentID int64
}

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

func (p PaymentID) String() string {
	return strconv.FormatInt(p.ProjectID, 10) + "-" + strconv.FormatInt(p.PaymentID, 10)
}

func (p PaymentID) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *PaymentID) UnmarshalJSON(data []byte) error {
	var str string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	*p, err = ParsePaymentIDStr(str)
	return err
}
