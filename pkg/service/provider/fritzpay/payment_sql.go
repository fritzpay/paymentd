package fritzpay

import (
	"database/sql"
	"errors"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
	"time"
)

var (
	ErrPaymentNotFound     = errors.New("payment not found")
	ErrTransactionNotFound = errors.New("transaction not found")
)

const selectPayment = `
SELECT
	id,
	project_id,
	payment_id,
	created,
	method_key
FROM provider_fritzpay_payment
`
const selectPaymentByPaymentID = selectPayment + `
WHERE
	project_id = ?
	AND
	payment_id = ?
`

func PaymentByPaymentIDTx(db *sql.Tx, id payment.PaymentID) (Payment, error) {
	row := db.QueryRow(selectPaymentByPaymentID, id.ProjectID, id.PaymentID)
	p := Payment{}
	err := row.Scan(
		&p.ID,
		&p.ProjectID,
		&p.PaymentID,
		&p.Created,
		&p.MethodKey,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrPaymentNotFound
		}
		return p, err
	}
	return p, nil
}

const insertPayment = `
INSERT INTO provider_fritzpay_payment
(project_id, payment_id, created, method_key)
VALUES
(?, ?, ?, ?)
`

func InsertPaymentTx(db *sql.Tx, p *Payment) error {
	stmt, err := db.Prepare(insertPayment)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(p.ProjectID, p.PaymentID, p.Created, p.MethodKey)
	stmt.Close()
	if err != nil {
		return err
	}
	p.ID, err = res.LastInsertId()
	return err
}

const insertPaymentTransaction = `
INSERT INTO provider_fritzpay_transaction
(fritzpay_payment_id, timestamp, status, fritzpay_id, payload)
VALUES
(?, ?, ?, ?, ?)
`

func InsertPaymentTransactionTx(db *sql.Tx, paymentTx PaymentTransaction) error {
	stmt, err := db.Prepare(insertPaymentTransaction)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		paymentTx.FritzpayPaymentID,
		paymentTx.Timestamp.UnixNano(),
		paymentTx.Status,
		paymentTx.FritzpayID,
		paymentTx.Payload,
	)
	stmt.Close()
	return err
}

const selectPaymentTransaction = `
SELECT
	t.fritzpay_payment_id,
	t.timestamp,
	t.status,
	t.fritzpay_id,
	t.payload
FROM provider_fritzpay_transaction AS t
`

const selectPaymentTransactionByID = selectPaymentTransaction + `
WHERE
	t.fritzpay_payment_id = ?
	AND
	t.timestamp = (
		SELECT MAX(timestamp) FROM provider_fritzpay_transaction
		WHERE
			fritzpay_payment_id = t.fritzpay_payment_id
	)
`

func PaymentTransactionCurrentByPaymentIDProviderTx(db *sql.Tx, id int64) (PaymentTransaction, error) {
	query := selectPaymentTransactionByID + `
AND
	t.status LIKE 'psp_%'
	`
	row := db.QueryRow(query, id)
	paymentTx := PaymentTransaction{}
	var ts int64
	err := row.Scan(
		&paymentTx.FritzpayPaymentID,
		&ts,
		&paymentTx.Status,
		&paymentTx.FritzpayID,
		&paymentTx.Payload,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return paymentTx, ErrTransactionNotFound
		}
		return paymentTx, err
	}
	paymentTx.Timestamp = time.Unix(0, ts)
	return paymentTx, nil
}
