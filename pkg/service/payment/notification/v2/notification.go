package notification

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/maputil"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"github.com/fritzpay/paymentd/pkg/service"
	"hash"
	"io"
	"strconv"
	"time"
)

const (
	PaymentNotificationVersion = "2.0.0-alpha"
)

// PaymentNotification represents a notification for connected systems about
// the state of a payment
type Notification struct {
	Version              string
	PaymentId            payment.PaymentID
	Ident                string
	Amount               int64 `json:",string"`
	Subunits             int8  `json:",string"`
	DecimalAmount        string
	Currency             string
	Country              string            `json:",omitempty"`
	PaymentMethodId      int64             `json:",string,omitempty"`
	Locale               string            `json:",omitempty"`
	Balance              payment.Balance   `json:",omitempty"`
	Status               string            `json:",omitempty"`
	TransactionTimestamp int64             `json:",string,omitempty"`
	Metadata             map[string]string `json:",omitempty"`
	Timestamp            int64             `json:",string"`
	Nonce                string            `json:",omitempty"`
	Signature            string            `json:",omitempty"`
}

func New(encodedPaymentID payment.PaymentID, p *payment.Payment) (*Notification, error) {
	n := &Notification{
		Version:       PaymentNotificationVersion,
		PaymentId:     encodedPaymentID,
		Ident:         p.Ident,
		Amount:        p.Amount,
		Subunits:      p.Subunits,
		DecimalAmount: p.Decimal().String(),
		Currency:      p.Currency,
		Status:        p.Status.String(),
		Metadata:      p.Metadata,
	}
	if !p.TransactionTimestamp.IsZero() {
		n.TransactionTimestamp = p.TransactionTimestamp.UnixNano()
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

func (n *Notification) Identification() string {
	return fmt.Sprintf("payment notification %s", n.Version)
}

func (n *Notification) SetTransactions(tl payment.PaymentTransactionList) {
	n.Balance = tl.Balance()
}

func (n *Notification) Sign(timestamp time.Time, nonce string, secret []byte) error {
	n.Timestamp = timestamp.Unix()
	n.Nonce = nonce
	sig, err := service.Sign(n, secret)
	if err != nil {
		return err
	}
	n.Signature = hex.EncodeToString(sig)
	return nil
}

func (n *Notification) Message() ([]byte, error) {
	var err error
	buf := bytes.NewBuffer(nil)
	_, err = buf.WriteString(n.Version)
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	_, err = buf.WriteString(n.PaymentId.String())
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	_, err = buf.WriteString(n.Ident)
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	_, err = buf.WriteString(strconv.FormatInt(n.Amount, 10))
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	_, err = buf.WriteString(strconv.FormatInt(int64(n.Subunits), 10))
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	_, err = buf.WriteString(n.DecimalAmount)
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	_, err = buf.WriteString(n.Currency)
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	_, err = buf.WriteString(n.Country)
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	if n.PaymentMethodId != 0 {
		_, err = buf.WriteString(strconv.FormatInt(int64(n.PaymentMethodId), 10))
		if err != nil {
			return nil, fmt.Errorf("buffer write error: %v", err)
		}
	}
	if n.Locale != "" {
		_, err = buf.WriteString(n.Locale)
		if err != nil {
			return nil, fmt.Errorf("buffer write error: %v", err)
		}
	}
	if n.Balance != nil {
		balanceMap := n.Balance.FlatMap()
		err = maputil.WriteSortedMap(buf, balanceMap)
		if err != nil {
			return nil, fmt.Errorf("buffer write error: %v", err)
		}
	}
	if n.Status != "" {
		_, err = buf.WriteString(n.Status)
		if err != nil {
			return nil, fmt.Errorf("buffer write error: %v", err)
		}
	}
	if n.TransactionTimestamp != 0 {
		_, err = buf.WriteString(strconv.FormatInt(int64(n.TransactionTimestamp), 10))
		if err != nil {
			return nil, fmt.Errorf("buffer write error: %v", err)
		}
	}
	if n.Metadata != nil {
		err = maputil.WriteSortedMap(buf, n.Metadata)
		if err != nil {
			return nil, fmt.Errorf("buffer write error: %v", err)
		}
	}
	_, err = buf.WriteString(strconv.FormatInt(int64(n.Timestamp), 10))
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	_, err = buf.WriteString(n.Nonce)
	if err != nil {
		return nil, fmt.Errorf("buffer write error: %v", err)
	}
	return buf.Bytes(), nil
}

func (n *Notification) HashFunc() func() hash.Hash {
	return sha256.New
}

func (n *Notification) Reader() io.ReadCloser {
	r, w := io.Pipe()
	go func() {
		enc := json.NewEncoder(w)
		err := enc.Encode(n)
		if err != nil {
			r.CloseWithError(err)
			w.CloseWithError(err)
			return
		}
		w.Close()
	}()
	return r
}
