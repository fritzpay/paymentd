package fritzpay

import (
	"database/sql"
	"errors"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
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
INSERT INTO provider_fritzpay_payment_transaction
(fritzpay_payment_id, timestamp, status, fritzpay_id, payload)
VALUES
(?, ?, ?, ?, ?)
`
