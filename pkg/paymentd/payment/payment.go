package payment

import (
	"code.google.com/p/godec/dec"
	"database/sql"
	"github.com/fritzpay/paymentd/pkg/decimal"
	"time"
)

const (
	// IdentMaxLen is the maximum allowed length of an identifier
	IdentMaxLen = 175
)

type Payment struct {
	projectID int64
	id        int64
	Created   time.Time
	Ident     string
	Amount    int64
	Subunits  int8
	Currency  string
	Country   string

	CallbackURL sql.NullString
	ReturnURL   sql.NullString
}

func (p *Payment) PaymentID() PaymentID {
	return PaymentID{p.ProjectID(), p.ID()}
}

func (p *Payment) ID() int64 {
	return p.id
}

func (p *Payment) ProjectID() int64 {
	return p.projectID
}

func (p *Payment) SetProjectID(projectID int64) {
	p.projectID = projectID
}

func (p *Payment) Decimal() *decimal.Decimal {
	d := dec.NewDecInt64(p.Amount)
	sc := dec.Scale(int32(p.Subunits))
	d.SetScale(sc)
	return &decimal.Decimal{Dec: d}
}