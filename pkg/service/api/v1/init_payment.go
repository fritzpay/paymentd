package v1

import (
	"bytes"
	"code.google.com/p/go.text/language"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/json"
	"github.com/fritzpay/paymentd/pkg/maputil"
	"github.com/fritzpay/paymentd/pkg/paymentd/nonce"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"gopkg.in/inconshreveable/log15.v2"
	"hash"
	"net/http"
	"net/url"
	"strconv"
	"unicode/utf8"
)

// InitPaymentRequest is the request JSON struct for POST /payment
type InitPaymentRequest struct {
	ProjectKey      string
	Ident           string
	Amount          json.RequiredInt64
	Subunits        json.RequiredInt8
	Currency        string
	Country         string
	PaymentMethodID int64  `json:"PaymentMethodId,string"`
	Locale          string `json:",omitempty"`
	CallbackURL     string `json:",omitempty"`
	ReturnURL       string `json:",omitempty"`

	Metadata map[string]string

	Timestamp int64 `json:",string"`
	Nonce     string

	HexSignature    string `json:"Signature"`
	binarySignature []byte
}

// Validate input
func (r *InitPaymentRequest) Validate() error {
	if r.ProjectKey == "" {
		return fmt.Errorf("missing ProjectKey")
	}
	if r.Ident == "" {
		return fmt.Errorf("missing Ident")
	}
	if utf8.RuneCountInString(r.Ident) > payment.IdentMaxLen {
		return fmt.Errorf("invalid Ident")
	}
	if !r.Amount.Set {
		return fmt.Errorf("missing Amount")
	}
	if r.Amount.Int64 < 0 {
		return fmt.Errorf("invalid Amount: %d", r.Amount.Int64)
	}
	if !r.Subunits.Set {
		return fmt.Errorf("missing Subunits")
	}
	if r.Currency == "" {
		return fmt.Errorf("missing Currency")
	}
	if len(r.Currency) != 3 {
		return fmt.Errorf("invalid Currency")
	}
	if r.Country == "" {
		return fmt.Errorf("missing Country")
	}
	if len(r.Country) != 2 {
		return fmt.Errorf("invalid Country")
	}
	if r.Timestamp == 0 {
		return fmt.Errorf("missing Timestamp")
	}
	if r.Nonce == "" {
		return fmt.Errorf("missing Nonce")
	}
	if len(r.Nonce) > nonce.NonceBytes {
		return fmt.Errorf("invalid Nonce")
	}
	if r.Locale != "" {
		if _, err := language.Parse(r.Locale); err != nil {
			return fmt.Errorf("invalid Locale")
		}
	}
	if r.CallbackURL != "" {
		if _, err := url.Parse(r.CallbackURL); err != nil {
			return fmt.Errorf("invalid CallbackURL")
		}
	}
	if r.ReturnURL != "" {
		if _, err := url.Parse(r.ReturnURL); err != nil {
			return fmt.Errorf("invalid ReturnURL")
		}
	}
	return nil
}

// Return the (binary) signature from the request
//
// implementing AuthenticatedRequest
func (r *InitPaymentRequest) Signature() ([]byte, error) {
	return hex.DecodeString(r.HexSignature)
}

// Message returns the signature base string as bytes or nil on error
func (r *InitPaymentRequest) Message() []byte {
	str, err := r.SignatureBaseString()
	if err != nil {
		return nil
	}
	return []byte(str)
}

// HashFunc returns the hash function used to generate a signature
func (r *InitPaymentRequest) HashFunc() func() hash.Hash {
	return sha256.New
}

// Return the signature base string (msg)
func (r *InitPaymentRequest) SignatureBaseString() (string, error) {
	var err error
	buf := bytes.NewBuffer(nil)
	_, err = buf.WriteString(r.ProjectKey)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Ident)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Amount.Int64, 10))
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(strconv.FormatInt(int64(r.Subunits.Int8), 10))
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Currency)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Country)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	if r.PaymentMethodID != 0 {
		_, err = buf.WriteString(strconv.FormatInt(r.PaymentMethodID, 10))
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.Locale != "" {
		_, err = buf.WriteString(r.Locale)
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.CallbackURL != "" {
		_, err = buf.WriteString(r.CallbackURL)
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.ReturnURL != "" {
		_, err = buf.WriteString(r.ReturnURL)
		if err != nil {
			return "", fmt.Errorf("buffer error: %v", err)
		}
	}
	if r.Metadata != nil {
		err = maputil.WriteSortedMap(buf, r.Metadata)
		if err != nil {
			return "", fmt.Errorf("error writing map: %v", err)
		}
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Timestamp, 10))
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	_, err = buf.WriteString(r.Nonce)
	if err != nil {
		return "", fmt.Errorf("buffer error: %v", err)
	}
	s := buf.String()
	return s, nil
}

// InitPaymentResponse is the JSON response struct for POST /payment
type InitPaymentResponse struct {
	Confirmation struct {
		Ident           string
		Amount          int64 `json:",string"`
		Subunits        int8  `json:",string"`
		Currency        string
		Country         string
		PaymentMethodId int64             `json:",string,omitempty"`
		Locale          string            `json:",omitempty"`
		CallbackURL     string            `json:",omitempty"`
		ReturnURL       string            `json:",omitempty"`
		Metadata        map[string]string `json:",omitempty"`
	}
	Payment struct {
		PaymentId payment.PaymentID
		// RFC3339 date/time string
		Created     string
		Token       string
		RedirectURL string
	}
	Timestamp int64 `json:",string"`
	Nonce     string
	Signature string
}

// ConfirmationFromPayment populates the response "Confirmation" object with
// the fields from the given payment
func (r *InitPaymentResponse) ConfirmationFromPayment(p payment.Payment) {
	r.Confirmation.Ident = p.Ident
	r.Confirmation.Amount = p.Amount
	r.Confirmation.Subunits = p.Subunits
	r.Confirmation.Currency = p.Currency
	r.Confirmation.Country = p.Country
}

// ConfirmationFromRequest populates the response "Confirmation" object with
// the fields from the given request
func (r *InitPaymentResponse) ConfirmationFromRequest(req *InitPaymentRequest) {
	if req.Locale != "" {
		r.Confirmation.Locale = req.Locale
	}
	if req.CallbackURL != "" {
		r.Confirmation.CallbackURL = req.CallbackURL
	}
	if req.ReturnURL != "" {
		r.Confirmation.ReturnURL = req.ReturnURL
	}
	if req.Metadata != nil {
		r.Confirmation.Metadata = req.Metadata
	}
}

// Returns the signature base string
//
// implementing SignableMessage
func (r *InitPaymentResponse) SignatureBaseString() string {
	var err error
	buf := bytes.NewBuffer(nil)
	_, err = buf.WriteString(r.Confirmation.Ident)
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Confirmation.Amount, 10))
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	_, err = buf.WriteString(strconv.FormatInt(int64(r.Confirmation.Subunits), 10))
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	_, err = buf.WriteString(r.Confirmation.Currency)
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	_, err = buf.WriteString(r.Confirmation.Country)
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	if r.Confirmation.PaymentMethodId != 0 {
		_, err = buf.WriteString(strconv.FormatInt(r.Confirmation.PaymentMethodId, 10))
		if err != nil {
			panic("buffer error: " + err.Error())
		}
	}
	if r.Confirmation.Locale != "" {
		_, err = buf.WriteString(r.Confirmation.Locale)
		if err != nil {
			panic("buffer error: " + err.Error())
		}
	}
	if r.Confirmation.CallbackURL != "" {
		_, err = buf.WriteString(r.Confirmation.CallbackURL)
		if err != nil {
			panic("buffer error: " + err.Error())
		}
	}
	if r.Confirmation.ReturnURL != "" {
		_, err = buf.WriteString(r.Confirmation.ReturnURL)
		if err != nil {
			panic("buffer error: " + err.Error())
		}
	}
	if r.Confirmation.Metadata != nil {
		maputil.WriteSortedMap(buf, r.Confirmation.Metadata)
	}
	_, err = buf.WriteString(r.Payment.PaymentId.String())
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	_, err = buf.WriteString(r.Payment.Created)
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	_, err = buf.WriteString(r.Payment.Token)
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	_, err = buf.WriteString(r.Payment.RedirectURL)
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	_, err = buf.WriteString(strconv.FormatInt(r.Timestamp, 10))
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	_, err = buf.WriteString(r.Nonce)
	if err != nil {
		panic("buffer error: " + err.Error())
	}
	s := buf.String()
	return s
}

func (a *PaymentAPI) InitPayment() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := a.log.New(log15.Ctx{
			"method": "InitPayment",
		})
		_ = log
	})
}
