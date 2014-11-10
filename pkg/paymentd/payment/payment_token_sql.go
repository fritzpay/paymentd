package payment

import (
	"database/sql"
	"time"

	"github.com/go-sql-driver/mysql"
)

func InsertPaymentTokenTx(tx *sql.Tx, t *PaymentToken) error {
	const insert = `
INSERT INTO payment_token
(token, created, project_id, payment_id)
VALUES
(?, ?, ?, ?)
`
	stmt, err := tx.Prepare(insert)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		t.Token,
		t.Created,
		t.id.ProjectID,
		t.id.PaymentID)
	stmt.Close()
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			// MySQL Error 1062 duplicate key
			if mysqlErr.Number == 1062 {
				err = t.GenerateToken()
				if err != nil {
					return err
				}
				return InsertPaymentTokenTx(tx, t)
			}
		}
		return err
	}
	return nil
}

const selectPaymentByToken = selectPaymentFields + `
FROM payment_token AS t
INNER JOIN payment AS p ON
	p.project_id = t.project_id
	AND
	p.id = t.payment_id
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
WHERE
	t.token = ?
	AND
	t.created > ?
`

func PaymentByTokenTx(db *sql.Tx, token string, tokenMaxAge time.Duration) (*Payment, error) {
	row := db.QueryRow(selectPaymentByToken, token, time.Now().Add(tokenMaxAge*-1))
	return scanSingleRow(row)
}

const deletePaymentToken = `
DELETE FROM payment_token WHERE token = ?
`

func DeletePaymentTokenTx(db *sql.Tx, token string) error {
	stmt, err := db.Prepare(deletePaymentToken)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(token)
	stmt.Close()
	return err
}
