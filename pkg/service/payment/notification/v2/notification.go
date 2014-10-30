package notification

import (
	"bytes"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
)

const (
	PaymentNotificationVersion = "2.0"
)

type PaymentNotification struct {
	Version              string
	PaymentId            payment.PaymentID
	Ident                string
	Amount               int64 `json:",string"`
	Subunits             int8  `json:",string"`
	DecimalAmount        string
	Currency             string
	Country              string          `json:",omitempty"`
	PaymentMethodId      int64           `json:",string,omitempty"`
	Locale               string          `json:",omitempty"`
	Balance              payment.Balance `json:",omitempty"`
	Status               string          `json:",omitempty"`
	TransactionTimestamp int64           `json:",string,omitempty"`
	Metadata             map[string]string
	Timestamp            int64  `json:",string"`
	Nonce                string `json:",omitempty"`
	Signature            string `json:",omitempty"`
}

func NewPaymentNotification(srv *paymentService.Service, p *payment.Payment) (*PaymentNotification, error) {
	n := &PaymentNotification{
		Version:       PaymentNotificationVersion,
		PaymentId:     srv.EncodedPaymentID(p.PaymentID()),
		Ident:         p.Ident,
		Amount:        p.Amount,
		Subunits:      p.Subunits,
		DecimalAmount: p.Decimal().String(),
		Currency:      p.Currency,
		Metadata:      p.Metadata,
	}
	if !p.Config.IsConfigured() {
		return n, nil
	}
	if p.Config.Country.Valid {
		n.Country = p.Config.Country.String
	}
	if p.Config.PaymentMethodID.Valid {
		n.PaymentMethodId = p.Config.PaymentMethodID.Int64
	}
	if p.Config.Locale.Valid {
		n.Locale = p.Config.Locale.String
	}
	return n, nil
}

func (n *PaymentNotification) Message() ([]byte, error) {
	var err error
	buf := bytes.NewBuffer(nil)
	_, err = buf.WriteString(n.Version)
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	return buf.Bytes(), nil
}
