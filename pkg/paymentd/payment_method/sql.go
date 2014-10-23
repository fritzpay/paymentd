package payment_method

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrPaymentMethodNotFound = errors.New("payment method not found")
)

const selectPaymentMethod = `
SELECT
	m.id
	m.project_id,
	p.id,
	p.name,
	m.method_key,
	m.created,
	m.created_by,
	s.status,
	s.timestamp,
	s.created_by
FROM payment_method AS m
INNER JOIN provider AS p ON
	p.id = m.provider_id
INNER JOIN payment_method_status AS s ON
	s.payment_method_id = m.id
	AND
	s.timestamp = (
		SELECT MAX(timestamp) FROM payment_method_status
		WHERE
			payment_method_id = s.payment_method_id
	)
`

const selectPaymentMethodByID = selectPaymentMethod + `
WHERE
	m.id = ?
`

func scanSinglePaymentMethod(row *sql.Row) (PaymentMethod, error) {
	pm := PaymentMethod{}
	var ts int64
	err := row.Scan(
		&pm.ID,
		&pm.ProjectID,
		&pm.Provider.ID,
		&pm.Provider.Name,
		&pm.MethodKey,
		&pm.Created,
		&pm.CreatedBy,
		&pm.Status,
		&ts,
		&pm.StatusCreatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return pm, ErrPaymentMethodNotFound
		}
		return pm, err
	}
	pm.StatusChanged = time.Unix(0, ts)
	return pm, nil
}

func PaymentMethodByIDDB(db *sql.DB, id int64) (PaymentMethod, error) {
	row := db.QueryRow(selectPaymentMethodByID, id)
	return scanSinglePaymentMethod(row)
}

func PaymentMethodByIDTx(db *sql.Tx, id int64) (PaymentMethod, error) {
	row := db.QueryRow(selectPaymentMethodByID, id)
	return scanSinglePaymentMethod(row)
}

const insertPaymentMethod = `
INSERT INTO payment_method
(project_id, provider_id, method_key, created, created_by)
VALUES
(?, ?, ?, ?, ?)
`
