package payment_method

import (
	"database/sql"
	"errors"
	"github.com/fritzpay/paymentd/pkg/metadata"
	"time"
)

var (
	ErrPaymentMethodNotFound  = errors.New("payment method not found")
	ErrPaymentMethodWithoutID = errors.New("payment method has no id")
)

const selectPaymentMethod = `
SELECT
	m.id,
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

func InsertPaymentMethodTx(db *sql.Tx, pm PaymentMethod) (int64, error) {
	stmt, err := db.Prepare(insertPaymentMethod)
	if err != nil {
		return 0, err
	}
	res, err := stmt.Exec(pm.ProjectID, pm.Provider.ID, pm.MethodKey, pm.Created, pm.CreatedBy)
	stmt.Close()
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

const insertPaymentMethodStatus = `
INSERT INTO payment_method_status
(payment_method_id, timestamp, status, created_by)
VALUES
(?, ?, ?, ?)
`

func InsertPaymentMethodStatusTx(db *sql.Tx, pm PaymentMethod) error {
	stmt, err := db.Prepare(insertPaymentMethodStatus)
	if err != nil {
		return err
	}
	ts := time.Now()
	_, err = stmt.Exec(pm.ID, ts.UnixNano(), pm.Status, pm.StatusCreatedBy)
	stmt.Close()
	return err
}

func InsertPaymentMethodMetadataTx(db *sql.Tx, pm PaymentMethod, createdBy string) error {
	if pm.ID == 0 {
		return ErrPaymentMethodWithoutID
	}
	m := metadata.MetadataFromValues(pm.Metadata, createdBy)
	return metadata.InsertMetadataTx(db, MetadataModel, pm.ID, m)
}
