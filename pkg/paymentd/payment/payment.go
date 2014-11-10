package payment

import (
	"database/sql"
	"fmt"
	"time"

	"code.google.com/p/godec/dec"
	"github.com/fritzpay/paymentd/pkg/decimal"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
)

const (
	// IdentMaxLen is the maximum allowed length of an identifier
	IdentMaxLen = 175
)

const (
	DefaultLocale = "en_US"
	// MetadataKeyAcceptLanguage is the key for the Accept-Language header
	// which will be stored in the metadata
	MetadataKeyAcceptLanguage = "_fAcceptLanguage"
	MetadataKeyBrowserLocale  = "_fBrowserLocale"
	MetadataKeyRemoteAddress  = "_fRemoteAddress"
)

// Payment represents a payment
type Payment struct {
	projectID int64
	id        int64
	Created   time.Time
	Ident     string
	Amount    int64
	Subunits  int8
	Currency  string

	Config Config

	TransactionTimestamp time.Time
	Status               PaymentTransactionStatus

	Metadata map[string]string
}

func (p *Payment) Valid() bool {
	return p.projectID != 0 && p.id != 0 && p.Ident != "" && p.Currency != ""
}

// PaymentID returns the identifier for the payment
func (p *Payment) PaymentID() PaymentID {
	return PaymentID{p.ProjectID(), p.ID()}
}

func (p *Payment) ID() int64 {
	return p.id
}

func (p *Payment) ProjectID() int64 {
	return p.projectID
}

func (p *Payment) SetProject(pr *project.Project) error {
	if pr.Empty() {
		return fmt.Errorf("cannot assign empty project")
	}
	p.projectID = pr.ID
	return nil
}

// Decimal returns the decimal representation of the Amount and Subunits values
func (p *Payment) Decimal() *decimal.Decimal {
	d := dec.NewDecInt64(p.Amount)
	sc := dec.Scale(int32(p.Subunits))
	d.SetScale(sc)
	return &decimal.Decimal{Dec: *d}
}

// HasTransaction returns true if the payment has a payment transaction entry
func (p *Payment) HasTransaction() bool {
	if p.TransactionTimestamp.IsZero() {
		return false
	}
	if !p.Status.Valid() {
		return false
	}
	return true
}

// NewTransaction creates a new payment transaction for this payment
//
// Its transaction fields will be populated with the copied values from the payment
func (p *Payment) NewTransaction(s PaymentTransactionStatus) *PaymentTransaction {
	return &PaymentTransaction{
		Payment: p,

		Timestamp: time.Now(),
		Amount:    p.Amount,
		Subunits:  p.Subunits,
		Currency:  p.Currency,
		Status:    s,
	}
}

type Config struct {
	Timestamp          time.Time
	PaymentMethodID    sql.NullInt64
	Country            sql.NullString
	Locale             sql.NullString
	CallbackURL        sql.NullString
	CallbackAPIVersion sql.NullString
	CallbackProjectKey sql.NullString
	ReturnURL          sql.NullString
}

func (cfg *Config) IsConfigured() bool {
	return cfg.PaymentMethodID.Valid && cfg.Country.Valid && cfg.Locale.Valid
}

func (cfg *Config) SetPaymentMethodID(id int64) {
	cfg.PaymentMethodID.Int64, cfg.PaymentMethodID.Valid = id, true
}

func (cfg *Config) SetCountry(country string) {
	cfg.Country.String, cfg.Country.Valid = country, true
}

func (cfg *Config) SetLocale(locale string) {
	cfg.Locale.String, cfg.Locale.Valid = locale, true
}

func (cfg *Config) SetCallbackURL(url string) {
	cfg.CallbackURL.String, cfg.CallbackURL.Valid = url, true
}

func (cfg *Config) SetCallbackAPIVersion(ver string) {
	cfg.CallbackAPIVersion.String, cfg.CallbackAPIVersion.Valid = ver, true
}

func (cfg *Config) SetCallbackProjectKey(key string) {
	cfg.CallbackProjectKey.String, cfg.CallbackProjectKey.Valid = key, true
}

func (cfg *Config) SetReturnURL(url string) {
	cfg.ReturnURL.String, cfg.ReturnURL.Valid = url, true
}

func (cfg *Config) HasCallback() bool {
	return cfg.CallbackURL.Valid && cfg.CallbackAPIVersion.Valid && cfg.CallbackProjectKey.Valid
}

func (cfg *Config) CallbackConfig() (url, apiVersion, projectKey string) {
	return cfg.CallbackURL.String, cfg.CallbackAPIVersion.String, cfg.CallbackProjectKey.String
}
