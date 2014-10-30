package v1

import (
	"bytes"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"strconv"
)

// GetPaymentRequest represents a get payment request
type GetPaymentRequest struct {
	ProjectKey string
	PaymentId  string
	paymentID  payment.PaymentID
	Ident      string
	Timestamp  int64
	Nonce      string
	Signature  string
}

func (r *GetPaymentRequest) SignatureBaseString() ([]byte, error) {
	var err error
	buf := bytes.NewBuffer(nil)
	_, err = buf.WriteString(r.ProjectKey)
	if err != nil {
		return nil, fmt.Errorf("buffer error: %v", err)
	}
	if r.PaymentId != "" {
		_, err = buf.WriteString(r.PaymentId)
		if err != nil {
			return nil, fmt.Errorf("buffer error: %v", err)
		}
	} else if r.Ident != "" {
		_, err = buf.WriteString(r.Ident)
		if err != nil {
			return nil, fmt.Errorf("buffer error: %v", err)
		}
	} else {
		return nil, fmt.Errorf("neither payment id nor ident set")
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Timestamp, 10))
	if err != nil {
		return nil, fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Nonce)
	if err != nil {
		return nil, fmt.Errorf("buffer error: %v", err)
	}
	return buf.Bytes(), nil
}

func (r *GetPaymentRequest) Message() []byte {
	msg, err := r.SignatureBaseString()
	if err != nil {
		return nil
	}
	return msg
}
