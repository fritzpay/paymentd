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

const selectPaymentFields = `
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
	c.callback_api_version,
	c.callback_project_key,
	c.return_url,
	c.expires,

	tx.timestamp,
	tx.status
`

const selectPayment = selectPaymentFields + `
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
		&p.Config.CallbackAPIVersion,
		&p.Config.CallbackProjectKey,
		&p.Config.ReturnURL,
		&p.Config.Expires,
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
		p.TransactionTimestamp = time.Unix(0, txTs.Int64)
	}
	return p, nil
}

func PaymentByIDTx(db *sql.Tx, id PaymentID) (*Payment, error) {
	row := db.QueryRow(selectPaymentByProjectIDAndID, id.ProjectID, id.PaymentID)
	return scanSingleRow(row)
}

func PaymentByIDDB(db *sql.DB, id PaymentID) (*Payment, error) {
	row := db.QueryRow(selectPaymentByProjectIDAndID, id.ProjectID, id.PaymentID)
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
(project_id, payment_id, timestamp, payment_method_id, country, locale, callback_url, callback_api_version, callback_project_key, return_url, expires)
VALUES
(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		p.Config.CallbackAPIVersion,
		p.Config.CallbackProjectKey,
		p.Config.ReturnURL,
		p.Config.Expires,
	)
	stmt.Close()
	return err
}

const selectPaymentMetadata = `
SELECT
	m.name,
	m.value
FROM payment_metadata AS m
WHERE
	m.project_id = ?
	AND
	m.payment_id = ?
	AND
	m.timestamp = (
		SELECT MAX(timestamp) FROM payment_metadata
		WHERE
			project_id = m.project_id
			AND
			payment_id = m.payment_id
	)
`

func scanPaymentMetadata(rows *sql.Rows, p *Payment) error {
	var err error
	meta := make(map[string]string)
	var k, v string
	for rows.Next() {
		err = rows.Scan(&k, &v)
		if err != nil {
			rows.Close()
			return err
		}
		meta[k] = v
	}
	p.Metadata = meta
	err = rows.Err()
	rows.Close()
	return err
}

func PaymentMetadataTx(db *sql.Tx, p *Payment) error {
	rows, err := db.Query(selectPaymentMetadata, p.ProjectID(), p.ID())
	if err != nil {
		return err
	}
	return scanPaymentMetadata(rows, p)
}

func PaymentMetadataDB(db *sql.DB, p *Payment) error {
	rows, err := db.Query(selectPaymentMetadata, p.ProjectID(), p.ID())
	if err != nil {
		return err
	}
	return scanPaymentMetadata(rows, p)
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
