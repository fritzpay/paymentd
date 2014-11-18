package paypal_rest

import (
	"database/sql"
	"errors"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
)

const selectTransaction = `
SELECT
	t.project_id,
	t.payment_id,
	t.timestamp,
	t.type,
	t.nonce,
	t.intent,
	t.paypal_id,
	t.payer_id,
	t.paypal_create_time,
	t.paypal_state,
	t.paypal_update_time,
	t.links,
	t.data
`

const selectTransactionCurrentByPaymentID = selectTransaction + `
FROM provider_paypal_transaction AS t
WHERE
	t.project_id = ?
	AND
	t.payment_id = ?
	AND
	t.timestamp = (
		SELECT MAX(timestamp) FROM provider_paypal_transaction
		WHERE
			project_id = t.project_id
			AND
			payment_id = t.payment_id
	)
`
const selectTransactionByPaymentIDAndNonce = selectTransaction + `
FROM provider_paypal_transaction AS tn
INNER JOIN provider_paypal_transaction AS t ON
	t.project_id = tn.project_id
	AND
	t.payment_id = tn.payment_id
	AND
	t.timestamp = (
		SELECT MAX(timestamp) FROM provider_paypal_transaction
		WHERE
			project_id = t.project_id
			AND
			payment_id = t.payment_id
	)
WHERE
	tn.project_id = ?
	AND
	tn.payment_id = ?
	AND
	tn.nonce = ?
`

func scanTransactionRow(row *sql.Row) (*Transaction, error) {
	t := &Transaction{}
	var ts int64
	err := row.Scan(
		&t.ProjectID,
		&t.PaymentID,
		&ts,
		&t.Type,
		&t.Nonce,
		&t.Intent,
		&t.PaypalID,
		&t.PayerID,
		&t.PaypalCreateTime,
		&t.PaypalState,
		&t.PaypalUpdateTime,
		&t.Links,
		&t.Data,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return t, ErrTransactionNotFound
		}
		return t, err
	}
	t.Timestamp = time.Unix(0, ts)
	return t, nil
}

func TransactionCurrentByPaymentIDTx(db *sql.Tx, paymentID payment.PaymentID) (*Transaction, error) {
	row := db.QueryRow(selectTransactionCurrentByPaymentID, paymentID.ProjectID, paymentID.PaymentID)
	return scanTransactionRow(row)
}

func TransactionCurrentByPaymentIDDB(db *sql.DB, paymentID payment.PaymentID) (*Transaction, error) {
	row := db.QueryRow(selectTransactionCurrentByPaymentID, paymentID.ProjectID, paymentID.PaymentID)
	return scanTransactionRow(row)
}

func TransactionByPaymentIDAndNonceTx(db *sql.Tx, paymentID payment.PaymentID, nonce string) (*Transaction, error) {
	row := db.QueryRow(selectTransactionByPaymentIDAndNonce, paymentID.ProjectID, paymentID.PaymentID, nonce)
	return scanTransactionRow(row)
}

func TransactionByPaymentIDAndNonceDB(db *sql.DB, paymentID payment.PaymentID, nonce string) (*Transaction, error) {
	row := db.QueryRow(selectTransactionByPaymentIDAndNonce, paymentID.ProjectID, paymentID.PaymentID, nonce)
	return scanTransactionRow(row)
}

const insertTransaction = `
INSERT INTO provider_paypal_transaction
(project_id, payment_id, timestamp, type, nonce, intent, paypal_id, payer_id, paypal_create_time, paypal_state, paypal_update_time, links, data)
VALUES
(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

func doInsertTransaction(stmt *sql.Stmt, t *Transaction) error {
	_, err := stmt.Exec(
		t.ProjectID,
		t.PaymentID,
		t.Timestamp.UnixNano(),
		t.Type,
		t.Nonce,
		t.Intent,
		t.PaypalID,
		t.PayerID,
		t.PaypalCreateTime,
		t.PaypalState,
		t.PaypalUpdateTime,
		t.Links,
		t.Data,
	)
	stmt.Close()
	return err
}

func InsertTransactionTx(db *sql.Tx, t *Transaction) error {
	stmt, err := db.Prepare(insertTransaction)
	if err != nil {
		return err
	}
	return doInsertTransaction(stmt, t)
}

func InsertTransactionDB(db *sql.DB, t *Transaction) error {
	stmt, err := db.Prepare(insertTransaction)
	if err != nil {
		return err
	}
	return doInsertTransaction(stmt, t)
}
