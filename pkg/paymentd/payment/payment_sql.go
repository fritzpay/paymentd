package payment

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

const insertPayment = `
INSERT INTO payment
(project_id, created, ident, amount, subunits, currency, callback_url, return_url)
VALUES
(?, ?, ?, ?, ?, ?, ?, ?)
`

func InsertPaymentTx(db *sql.Tx, p *Payment) error {
	stmt, err := db.Prepare(insertPayment)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(
		p.ProjectID(),
		p.Created,
		p.Ident,
		p.Amount,
		p.Subunits,
		p.Currency,
		p.CallbackURL,
		p.ReturnURL,
	)
	stmt.Close()
	if err != nil {
		return err
	}
	p.id, err = res.LastInsertId()
	return err
}

const selectPayment = `
SELECT
	id,
	project_id,
	created,
	ident,
	amount,
	subunits,
	currency,
	callback_url,
	return_url
FROM payment
`

const selectPaymentByProjectIDAndIdent = selectPayment + `
WHERE
	project_id = ?
	AND
	ident = ?
`

func scanSingleRow(row *sql.Row) (*Payment, error) {
	p := &Payment{}
	err := row.Scan(
		&p.id,
		&p.projectID,
		&p.Created,
		&p.Ident,
		&p.Amount,
		&p.Subunits,
		&p.Currency,
		&p.CallbackURL,
		&p.ReturnURL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrPaymentNotFound
		}
		return p, err
	}
	return p, nil
}

func PaymentByProjectIDAndIdentTx(db *sql.Tx, projectID int64, ident string) (*Payment, error) {
	row := db.QueryRow(selectPaymentByProjectIDAndIdent, projectID, ident)
	return scanSingleRow(row)
}

const insertPaymentConfig = `
INSERT INTO payment_config
(project_id, payment_id, timestamp, payment_method_id, country, locale)
VALUES
(?, ?, ?, ?, ?, ?)
`

func InsertPaymentConfigTx(db *sql.Tx, p *Payment) error {
	stmt, err := db.Prepare(insertPaymentConfig)
	if err != nil {
		return err
	}
	ts := time.Now().UnixNano()
	_, err = stmt.Exec(
		p.ProjectID(),
		p.ID(),
		ts,
		p.Config.PaymentMethodID,
		p.Config.Country,
		p.Config.Locale,
	)
	stmt.Close()
	return err
}

const insertPaymentMetadata = `
INSERT INTO payment_metadata
(project_id, payment_id, name, timestamp, value)
VALUES
(?, ?, ?, ?, ?)
`

func InsertPaymentMetadataTx(db *sql.Tx, p *Payment) error {
	if p.Metadata == nil {
		return nil
	}
	stmt, err := db.Prepare(insertPaymentMetadata)
	if err != nil {
		return err
	}
	ts := time.Now().UnixNano()
	for n, v := range p.Metadata {
		_, err = stmt.Exec(
			p.ProjectID(),
			p.ID(),
			n,
			ts,
			v,
		)
		if err != nil {
			stmt.Close()
			return err
		}
	}
	stmt.Close()
	return nil
}
