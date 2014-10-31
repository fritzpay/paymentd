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
(project_id, created, ident, amount, subunits, currency)
VALUES
(?, ?, ?, ?, ?, ?)
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
	p.id,
	p.project_id,
	p.created,
	p.ident,
	p.amount,
	p.subunits,
	p.currency,

	c.timestamp,
	c.payment_method_id,
	c.country,
	c.locale,
	c.callback_url,
	c.return_url,

	tx.timestamp,
	tx.status
FROM payment AS p
LEFT JOIN payment_config AS c ON
	c.project_id = p.project_id
	AND
	c.payment_id = p.id
	AND
	c.timestamp = (
		SELECT MAX(timestamp) FROM payment_config
		WHERE
			project_id = c.project_id
			AND
			payment_id = c.payment_id
	)
LEFT JOIN payment_transaction AS tx ON
	tx.project_id = p.project_id
	AND
	tx.payment_id = p.id
	AND
	tx.timestamp = (
		SELECT MAX(timestamp) FROM payment_transaction
		WHERE
			project_id = tx.project_id
			AND
			payment_id = tx.payment_id
	)
`

const selectPaymentByProjectIDAndID = selectPayment + `
WHERE
	p.project_id = ?
	AND
	p.id = ?
`

const selectPaymentByProjectIDAndIdent = selectPayment + `
WHERE
	p.project_id = ?
	AND
	p.ident = ?
`

func scanSingleRow(row *sql.Row) (*Payment, error) {
	p := &Payment{}
	var ts, txTs sql.NullInt64
	err := row.Scan(
		&p.id,
		&p.projectID,
		&p.Created,
		&p.Ident,
		&p.Amount,
		&p.Subunits,
		&p.Currency,
		&ts,
		&p.Config.PaymentMethodID,
		&p.Config.Country,
		&p.Config.Locale,
		&p.Config.CallbackURL,
		&p.Config.ReturnURL,
		&txTs,
		&p.Status,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrPaymentNotFound
		}
		return p, err
	}
	if ts.Valid {
		p.Config.Timestamp = time.Unix(0, ts.Int64)
	}
	if txTs.Valid {
		p.TransactionTimestamp = time.Unix(0, ts.Int64)
	}
	return p, nil
}

func PaymentByProjectIDAndIDDB(db *sql.DB, projectID int64, id int64) (*Payment, error) {
	row := db.QueryRow(selectPaymentByProjectIDAndID, projectID, id)
	return scanSingleRow(row)
}

func PaymentByProjectIDAndIdentDB(db *sql.DB, projectID int64, ident string) (*Payment, error) {
	row := db.QueryRow(selectPaymentByProjectIDAndIdent, projectID, ident)
	return scanSingleRow(row)
}

func PaymentByProjectIDAndIdentTx(db *sql.Tx, projectID int64, ident string) (*Payment, error) {
	row := db.QueryRow(selectPaymentByProjectIDAndIdent, projectID, ident)
	return scanSingleRow(row)
}

const insertPaymentConfig = `
INSERT INTO payment_config
(project_id, payment_id, timestamp, payment_method_id, country, locale, callback_url, return_url)
VALUES
(?, ?, ?, ?, ?, ?, ?, ?)
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
		p.Config.CallbackURL,
		p.Config.ReturnURL,
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
