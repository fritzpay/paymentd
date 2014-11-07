package payment

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrPaymentTransactionNotFound = errors.New("payment transaction not found")
)

const selectPaymentTransaction = `
SELECT
	tx.timestamp,

	tx.amount,
	tx.subunits,
	tx.currency,
	tx.status,
	tx.comment
FROM payment_transaction AS tx
`

const insertPaymentTransaction = `
INSERT INTO payment_transaction
(project_id, payment_id, timestamp, amount, subunits, currency, status, comment)
VALUES
(?, ?, ?, ?, ?, ?, ?, ?)
`

func InsertPaymentTransactionTx(db *sql.Tx, paymentTx *PaymentTransaction) error {
	stmt, err := db.Prepare(insertPaymentTransaction)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		paymentTx.Payment.ProjectID(),
		paymentTx.Payment.ID(),
		paymentTx.Timestamp.UnixNano(),
		paymentTx.Amount,
		paymentTx.Subunits,
		paymentTx.Currency,
		paymentTx.Status,
		paymentTx.Comment,
	)
	stmt.Close()
	return err
}

const selectCurrentPaymentTransaction = selectPaymentTransaction + `
WHERE
	tx.project_id = ?
	AND
	tx.payment_id = ?
	AND
	tx.timestamp = (
		SELECT MAX(timestamp) FROM payment_transaction
		WHERE
			project_id = tx.project_id
			AND
			payment_id = tx.payment_id
	)
`

// PaymentTransactionCurrentTx reads the current payment transaction into the given payment
// (i.e. it sets the TransactionTimestamp and Status fields) and returns the full
// PaymentTransaction type
//
// If no payment transaction exists, it will return an ErrPaymentTransactionNotFound
func PaymentTransactionCurrentTx(db *sql.Tx, p *Payment) (*PaymentTransaction, error) {
	paymentTx := &PaymentTransaction{
		Payment: p,
	}
	var ts int64
	row := db.QueryRow(selectCurrentPaymentTransaction, p.ProjectID(), p.ID())
	err := row.Scan(
		&ts,
		&paymentTx.Amount,
		&paymentTx.Subunits,
		&paymentTx.Currency,
		&paymentTx.Status,
		&paymentTx.Comment,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return paymentTx, ErrPaymentTransactionNotFound
		}
		return paymentTx, err
	}
	paymentTx.Timestamp = time.Unix(0, ts)
	p.TransactionTimestamp = paymentTx.Timestamp
	p.Status = paymentTx.Status
	return paymentTx, nil
}
