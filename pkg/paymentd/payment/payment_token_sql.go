package payment

import (
	"database/sql"
	"github.com/go-sql-driver/mysql"
)

func InsertPaymentTokenTx(tx *sql.Tx, t PaymentToken) error {
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
			// MySQL Error 1068 SQLSTATE: 42000 (ER_MULTIPLE_PRI_KEY)
			if mysqlErr.Number == 1068 {
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
