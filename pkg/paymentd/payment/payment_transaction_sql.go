package payment

import (
	"database/sql"
)

const selectPaymentTransaction = `
SELECT
	tx.project_id,
	tx.payment_id,
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
		paymentTx.Amount,
		paymentTx.Subunits,
		paymentTx.Currency,
		paymentTx.Status,
		paymentTx.Comment,
	)
	stmt.Close()
	return err
}
