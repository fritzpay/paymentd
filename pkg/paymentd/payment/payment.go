package payment

import (
	"code.google.com/p/godec/dec"
	"database/sql"
	"fmt"
	"github.com/fritzpay/paymentd/pkg/decimal"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"time"
)

const (
	// IdentMaxLen is the maximum allowed length of an identifier
	IdentMaxLen = 175
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
func (p *Payment) Decimal() decimal.Decimal {
	d := dec.NewDecInt64(p.Amount)
	sc := dec.Scale(int32(p.Subunits))
	d.SetScale(sc)
	return decimal.Decimal{Dec: d}
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
	Timestamp       time.Time
	PaymentMethodID sql.NullInt64
	Country         sql.NullString
	Locale          sql.NullString
	CallbackURL     sql.NullString
	ReturnURL       sql.NullString
}

func (cfg *Config) IsConfigured() bool {
	return cfg.PaymentMethodID.Valid && cfg.Country.Valid && cfg.Locale.Valid
}
