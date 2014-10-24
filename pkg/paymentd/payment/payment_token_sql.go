package payment

import (
	"database/sql"
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
