// Paypal data types
package paypal_rest

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/nonce"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	paymentIDParam         = "paymentID"
	nonceParam             = "nonce"
	paypalPayerIDParameter = "PayerID"
)

type PayPalPaymentMethod string

const (
	PayPalPaymentMethodPayPal PayPalPaymentMethod = "paypal"
	PayPalPaymentMethodCC                         = "credit_card"
)

const (
	IntentSale = "sale"
	IntentAuth = "authorize"
)

// PayPal transaction types
const (
	TransactionTypeCreatePayment          = "createPayment"
	TransactionTypeCreatePaymentResponse  = "createPaymentResponse"
	TransactionTypeError                  = "error"
	TransactionTypeExecutePayment         = "executePayment"
	TransactionTypeExecutePaymentResponse = "executePaymentResponse"
	TransactionTypeGetPayment             = "getPayment"
	TransactionTypeGetPaymentResponse     = "getPaymentResponse"
)

var (
	ErrNoLinks           = errors.New("no links")
	ErrPayPalPaymentNoID = errors.New("paypal payment withoud ID")
)

type PayPalError struct {
	Name            string `json:"name"`
	Message         string `json:"message"`
	InformationLink string `json:"information_link"`
	Details         string `json:"details"`
}

// PaypalPayer represents the "payer" object as defined by the PayPal REST-API
//
// See https://developer.paypal.com/docs/api/#payer-object
type PaypalPayer struct {
	PaymentMethod PayPalPaymentMethod `json:"payment_method"`
	Status        string              `json:"status,omitempty"`
}

type PayPalPayerInfo struct {
	Email           string                `json:"email,omitempty"`
	FirstName       string                `json:"first_name,omitempty"`
	LastName        string                `json:"last_name,omitempty"`
	PayerID         string                `json:"payer_id,omitempty"`
	Phone           string                `json:"phone,omitempty"`
	ShippingAddress PayPalShippingAddress `json:"shipping_address,omitempty"`
	TaxIDType       string                `json:"tax_id_type,omitempty"`
	TaxID           string                `json:"tax_id,omitempty"`
}

type PayPalShippingAddress struct {
	RecipientName string `json:"recipient_name,omitempty"`
	Type          string `json:"type,omitempty"`
	Line1         string `json:"line1"`
	Line2         string `json:"line2,omitempty"`
	City          string `json:"city"`
	CountryCode   string `json:"country_code"`
	PostalCode    string `json:"postal_code,omitempty"`
	State         string `json:"state,omitempty"`
	Phone         string `json:"phone,omitempty"`
}

// PayPalDetails represents the PayPal "amount" details type
//
// See https://developer.paypal.com/docs/api/#details-object
type PayPalDetails struct {
	Shipping         string `json:"shipping,omitempty"`
	Subtotal         string `json:"subtotal,omitempty"`
	Tax              string `json:"tax,omitempty"`
	Fee              string `json:"fee,omitempty"`
	HandlingFee      string `json:"handling_fee,omitempty"`
	Insurance        string `json:"insurance,omitempty"`
	ShippingDiscount string `json:"shipping_discount,omitempty"`
}

type PayPalAmount struct {
	Currency string         `json:"currency"`
	Total    string         `json:"total"`
	Details  *PayPalDetails `json:"details,omitempty"`
}

type PayPalTransaction struct {
	Amount           PayPalAmount    `json:"amount"`
	Description      string          `json:"description,omitempty"`
	RelatedResources PayPalResources `json:"related_resources,omitempty"`
	InvoiceNumber    string          `json:"invoice_number,omitempty"`
	Custom           string          `json:"custom,omitempty"`
	SoftDescriptor   string          `json:"soft_descriptor,omitempty"`
}

type PayPalRedirectURLs struct {
	ReturnURL string `json:"return_url"`
	CancelURL string `json:"cancel_url"`
}

type PayPalLink struct {
	HRef   string `json:"href"`
	Rel    string `json:"rel"`
	Method string `json:"method"`
}

type PayPalPaymentRequest struct {
	Intent       string              `json:"intent"`
	Payer        PaypalPayer         `json:"payer"`
	Transactions []PayPalTransaction `json:"transactions"`
	RedirectURLs PayPalRedirectURLs  `json:"redirect_urls,omitempty"`
}

type PaypalPayment struct {
	PayPalPaymentRequest

	ID         string       `json:"id"`
	CreateTime string       `json:"create_time"`
	State      string       `json:"state"`
	UpdateTime string       `json:"update_time"`
	Links      []PayPalLink `json:"links"`
}

type PayPalPaymentExecution struct {
	PayerID      string              `json:"payer_id"`
	Transactions []PayPalTransaction `json:"transactions,omitempty"`
}

type PayPalResources []map[string]PayPalResource

func (p PayPalResources) Resources(t string) []PayPalResource {
	resources := make([]PayPalResource, 0, len(p))
	for _, m := range p {
		if r, ok := m[t]; ok {
			resources = append(resources, r)
		}
	}
	return resources
}

// PayPalResource represents one of sale, authorization, capture or refund object
type PayPalResource struct {
	ID                        string       `json:"id"`
	Amount                    PayPalAmount `json:"amount"`
	IsFinalCapture            bool         `json:"is_final_capture"`
	Description               string       `json:"string,omitempty"`
	CreateTime                string       `json:"create_time"`
	State                     string       `json:"state"`
	CaptureID                 string       `json:"capture_id,omitempty"`
	ParentPayment             string       `json:"parent_payment"`
	ValidUntil                string       `json:"valid_until,omitempty"`
	UpdateTime                string       `json:"update_time"`
	PaymentMode               string       `json:"payment_mode,omitempty"`
	PendingReason             string       `json:"pending_reason,omitempty"`
	ReasonCode                string       `json:"reason_code,omitempty"`
	ClearingTime              string       `json:"clearing_time,omitempty"`
	ProtectionEligibility     string       `json:"protection_eligibility,omitempty"`
	ProtectionEligibilityType string       `json:"protection_eligibility_type,omitempty"`
	Links                     []PayPalLink `json:"links,omitempty"`
}

func (d *Driver) createPaypalPaymentRequest(p *payment.Payment, cfg *Config, non *nonce.Nonce) (*PayPalPaymentRequest, error) {
	if cfg.Type != IntentSale && cfg.Type != IntentAuth {
		return nil, fmt.Errorf("invalid config. type %s not recognized", cfg.Type)
	}
	var err error
	req := &PayPalPaymentRequest{}
	req.Intent = cfg.Type
	req.Payer.PaymentMethod = PayPalPaymentMethodPayPal
	req.RedirectURLs, err = d.redirectURLs(p, urlSetNonce(non.Nonce))
	if err != nil {
		d.log.Error("error creating redirect urls", log15.Ctx{"err": err})
		return nil, ErrInternal
	}
	req.Transactions = []PayPalTransaction{
		d.payPalTransactionFromPayment(p),
	}
	return req, nil
}

func (d *Driver) payPalTransactionFromPayment(p *payment.Payment) PayPalTransaction {
	t := PayPalTransaction{}
	encPaymentID := d.paymentService.EncodedPaymentID(p.PaymentID())
	t.Custom = encPaymentID.String()
	t.InvoiceNumber = encPaymentID.String()
	t.Amount = PayPalAmount{
		Currency: p.Currency,
		Total:    p.DecimalRound(2).String(),
	}
	return t
}

type urlModification func(u *url.URL) error

var urlSetNonce = func(nonce string) urlModification {
	return urlModification(func(u *url.URL) error {
		q := u.Query()
		q.Set("nonce", nonce)
		u.RawQuery = q.Encode()
		return nil
	})
}

func (d *Driver) redirectURLs(p *payment.Payment, mods ...urlModification) (PayPalRedirectURLs, error) {
	u := PayPalRedirectURLs{}
	returnRoute, err := d.mux.Get("returnHandler").URLPath()
	if err != nil {
		return u, err
	}
	cancelRoute, err := d.mux.Get("cancelHandler").URLPath()
	if err != nil {
		return u, err
	}

	q := url.Values(make(map[string][]string))
	q.Set(paymentIDParam, d.paymentService.EncodedPaymentID(p.PaymentID()).String())

	returnURL, err := d.baseURL()
	if err != nil {
		return u, err
	}
	returnURL.Path = returnRoute.Path
	returnURL.RawQuery = q.Encode()

	cancelURL, err := d.baseURL()
	if err != nil {
		return u, err
	}
	cancelURL.Path = cancelRoute.Path
	cancelURL.RawQuery = q.Encode()

	for _, mod := range mods {
		err = mod(returnURL)
		if err != nil {
			return u, err
		}
		err = mod(cancelURL)
		if err != nil {
			return u, err
		}
	}

	u.ReturnURL = returnURL.String()
	u.CancelURL = cancelURL.String()

	return u, nil
}

type Config struct {
	ProjectID int64
	MethodKey string
	Created   time.Time
	CreatedBy string

	Endpoint string
	ClientID string
	Secret   string
	Type     string
}

// Transaction represents a transaction on a paypal payment
//
// It can be one of the following:
//
//   - A representation of a request.
//   - A representation of a response.
//   - A representation of a local change to a transaction.
//
// It also keeps the state of the paypal payment, i.e. the most recent transaction
// will denote what the state of the payment is.
type Transaction struct {
	ProjectID        int64
	PaymentID        int64
	Timestamp        time.Time
	Type             string
	Nonce            sql.NullString
	Intent           sql.NullString
	PaypalID         sql.NullString
	PayerID          sql.NullString
	PaypalCreateTime *time.Time
	PaypalState      sql.NullString
	PaypalUpdateTime *time.Time
	Links            []byte
	Data             []byte
}

func (t *Transaction) SetNonce(nonce string) {
	t.Nonce.String, t.Nonce.Valid = nonce, true
}

func (t *Transaction) SetIntent(intent string) {
	t.Intent.String, t.Intent.Valid = intent, true
}

func (t *Transaction) SetPaypalID(id string) {
	t.PaypalID.String, t.PaypalID.Valid = id, true
}

func (t *Transaction) SetPayerID(id string) {
	t.PayerID.String, t.PayerID.Valid = id, true
}

func (t *Transaction) SetState(state string) {
	t.PaypalState.String, t.PaypalState.Valid = state, true
}

func (t *Transaction) PayPalLinks() (map[string]*PayPalLink, error) {
	if t.Links == nil || len(t.Links) == 0 {
		return nil, ErrNoLinks
	}
	links := make([]*PayPalLink, 0, 10)
	err := json.Unmarshal(t.Links, &links)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]*PayPalLink)
	for _, l := range links {
		ret[l.Rel] = l
	}
	return ret, nil
}

func NewPayPalPaymentTransaction(paypalP *PaypalPayment) (*Transaction, error) {
	var err error
	paypalTx := &Transaction{
		Timestamp: time.Now(),
	}
	if paypalP.Intent != "" {
		paypalTx.SetIntent(paypalP.Intent)
	}
	if paypalP.ID != "" {
		paypalTx.SetPaypalID(paypalP.ID)
	}
	if paypalP.State != "" {
		paypalTx.SetState(paypalP.State)
	}
	if paypalP.CreateTime != "" {
		var t time.Time
		t, err = time.Parse(time.RFC3339, paypalP.CreateTime)
		if err == nil {
			paypalTx.PaypalCreateTime = &t
		}
	}
	if paypalP.UpdateTime != "" {
		var t time.Time
		t, err = time.Parse(time.RFC3339, paypalP.UpdateTime)
		if err == nil {
			paypalTx.PaypalUpdateTime = &t
		}
	}
	paypalTx.Links, err = json.Marshal(paypalP.Links)
	if err != nil {
		return nil, err
	}
	paypalTx.Data, err = json.Marshal(paypalP)
	if err != nil {
		return nil, err
	}
	return paypalTx, err
}

type Authorization struct {
	ProjectID       int64
	PaymentID       int64
	Timestamp       time.Time
	ValidUntil      time.Time
	State           string
	AuthorizationID string
	PaypalID        string
	Amount          string
	Currency        string
	Links           []byte
	Data            []byte
}

// NewPayPalPaymentAuthorization creates an authorization entry for the given payment
// and PayPal payment type
//
// The PayPal documentation is lacking information about how multiple transactions
// are handled. We will try a somewhat naïve approach here.
func NewPayPalPaymentAuthorization(p *payment.Payment, paypalP *PaypalPayment) (*Authorization, error) {
	if paypalP.ID == "" {
		return nil, ErrPayPalPaymentNoID
	}
	auth := &Authorization{
		ProjectID: p.ProjectID(),
		PaymentID: p.ID(),
		Timestamp: time.Now(),
		PaypalID:  paypalP.ID,
	}
	var authRes *PayPalResource
	for _, tx := range paypalP.Transactions {
		authLen := len(tx.RelatedResources.Resources("authorization"))
		if authLen == 0 {
			continue
		}
		// we assume this is illegal
		if authLen > 1 {
			return nil, fmt.Errorf("multiple authorizations in related resources")
		}
		authRes = &tx.RelatedResources.Resources("authorization")[0]
	}
	if authRes == nil {
		return nil, fmt.Errorf("no authorization resource")
	}
	valid, err := time.Parse(time.RFC3339, authRes.ValidUntil)
	if err != nil {
		return nil, fmt.Errorf("error parsing validity: %v", err)
	}
	auth.ValidUntil = valid
	auth.State = authRes.State
	auth.AuthorizationID = authRes.ID
	auth.Amount = authRes.Amount.Total
	auth.Currency = authRes.Amount.Currency
	if authRes.Links != nil {
		links, err := json.Marshal(authRes.Links)
		if err != nil {
			return nil, fmt.Errorf("error encoding links: %v", err)
		}
		auth.Links = links
	}
	enc, err := json.Marshal(authRes)
	if err != nil {
		return nil, fmt.Errorf("error encoding authorization: %v", err)
	}
	auth.Data = enc
	return auth, nil
}
