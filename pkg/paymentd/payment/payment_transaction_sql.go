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

type resultScanner interface {
	Scan(...interface{}) error
}

func scanPaymentTx(r resultScanner, paymentTx *PaymentTransaction) error {
	var ts int64
	err := r.Scan(
		&ts,
		&paymentTx.Amount,
		&paymentTx.Subunits,
		&paymentTx.Currency,
		&paymentTx.Status,
		&paymentTx.Comment,
	)
	paymentTx.Timestamp = time.Unix(0, ts)
	return err
}

// PaymentTransactionCurrentTx reads the current payment transaction into the given payment
// (i.e. it sets the TransactionTimestamp and Status fields) and returns the full
// PaymentTransaction type
//
// If no payment transaction exists, it will return an ErrPaymentTransactionNotFound
func PaymentTransactionCurrentTx(db *sql.Tx, p *Payment) (*PaymentTransaction, error) {
	paymentTx := &PaymentTransaction{
		Payment: p,
	}
	row := db.QueryRow(selectCurrentPaymentTransaction, p.ProjectID(), p.ID())
	err := scanPaymentTx(row, paymentTx)
	if err != nil {
		if err == sql.ErrNoRows {
			return paymentTx, ErrPaymentTransactionNotFound
		}
		return paymentTx, err
	}
	p.TransactionTimestamp = paymentTx.Timestamp
	p.Status = paymentTx.Status
	return paymentTx, nil
}

const selectPaymentTransactionsBefore = selectPaymentTransaction + `
WHERE
	tx.project_id = ?
	AND
	tx.payment_id = ?
	AND
	tx.timestamp <= ?
ORDER BY tx.timestamp ASC
`

func scanTransactions(rows *sql.Rows, p *Payment) (PaymentTransactionList, error) {
	var err error
	txs := make([]*PaymentTransaction, 0, 1)
	for rows.Next() {
		tx := &PaymentTransaction{Payment: p}
		err = scanPaymentTx(rows, tx)
		if err != nil {
			rows.Close()
			return nil, err
		}
		txs = append(txs, tx)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if len(txs) == 0 {
		return nil, ErrPaymentTransactionNotFound
	}
	return PaymentTransactionList(txs), nil
}

// PaymentTransactionsBeforeDB returns a PaymentTransactionList with all transactions
// before and including the given payment transaction.
//
// The list will be sorted by the earliest tx first.
func PaymentTransactionsBeforeDB(db *sql.DB, paymentTx *PaymentTransaction) (PaymentTransactionList, error) {
	query, err := db.Query(
		selectPaymentTransactionsBefore,
		paymentTx.Payment.ProjectID(),
		paymentTx.Payment.ID(),
		paymentTx.Timestamp.UnixNano(),
	)
	if err != nil {
		return nil, err
	}
	return scanTransactions(query, paymentTx.Payment)
}

// PaymentTransactionsBeforeDB returns a PaymentTransactionList with all transactions
// before and including the given payment transaction.
//
// The list will be sorted by the earliest tx first.
func PaymentTransactionsBeforeTimestampDB(db *sql.DB, p *Payment, transactionTimestamp time.Time) (PaymentTransactionList, error) {
	query, err := db.Query(
		selectPaymentTransactionsBefore,
		p.ProjectID(),
		p.ID(),
		transactionTimestamp.UnixNano(),
	)
	if err != nil {
		return nil, err
	}
	return scanTransactions(query, p)
}
