package payment_method

import (
	"database/sql"
	"errors"
	"time"

	"github.com/fritzpay/paymentd/pkg/metadata"
)

var (
	ErrPaymentMethodNotFound  = errors.New("payment method not found")
	ErrPaymentMethodWithoutID = errors.New("payment method has no id")
)

const selectPaymentMethod = `
SELECT
	m.id,
	m.project_id,
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
const selectPaymentMethodByProjectIDProviderIDMethodKey = selectPaymentMethod + `
WHERE
	m.project_id = ?
AND
	p.name = ?
AND
	m.method_key = ?
`

func scanSinglePaymentMethod(row *sql.Row) (*Method, error) {
	pm := &Method{}
	var ts int64
	err := row.Scan(
		&pm.ID,
		&pm.ProjectID,
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

func PaymentMethodByIDDB(db *sql.DB, id int64) (*Method, error) {
	row := db.QueryRow(selectPaymentMethodByID, id)
	return scanSinglePaymentMethod(row)
}

func PaymentMethodByProjectIDProviderNameMethodKeyDB(db *sql.DB, project_id int64, provider string, method_key string) (*Method, error) {
	row := db.QueryRow(selectPaymentMethodByProjectIDProviderIDMethodKey, project_id, provider, method_key)
	return scanSinglePaymentMethod(row)
}

func PaymentMethodByProjectIDProviderNameMethodKeyTx(tx *sql.Tx, project_id int64, provider string, method_key string) (*Method, error) {
	row := tx.QueryRow(selectPaymentMethodByProjectIDProviderIDMethodKey, project_id, provider, method_key)
	return scanSinglePaymentMethod(row)
}

func PaymentMethodByIDTx(db *sql.Tx, id int64) (*Method, error) {
	row := db.QueryRow(selectPaymentMethodByID, id)
	return scanSinglePaymentMethod(row)
}

const insertPaymentMethod = `
INSERT INTO payment_method
(project_id, provider, method_key, created, created_by)
VALUES
(?, ?, ?, ?, ?)
`

func InsertPaymentMethodTx(db *sql.Tx, pm *Method) error {
	stmt, err := db.Prepare(insertPaymentMethod)
	if err != nil {
		return err
	}
	res, err := stmt.Exec(pm.ProjectID, pm.Provider.Name, pm.MethodKey, pm.Created, pm.CreatedBy)
	stmt.Close()
	if err != nil {
		return err
	}
	pm.ID, err = res.LastInsertId()
	return err
}

const insertPaymentMethodStatus = `
INSERT INTO payment_method_status
(payment_method_id, timestamp, status, created_by)
VALUES
(?, ?, ?, ?)`

func InsertPaymentMethodStatusTx(db *sql.Tx, pm *Method) error {
	stmt, err := db.Prepare(insertPaymentMethodStatus)
	if err != nil {
		return err
	}
	ts := time.Now()
	_, err = stmt.Exec(pm.ID, ts.UnixNano(), pm.Status, pm.StatusCreatedBy)
	stmt.Close()
	return err
}

func InsertPaymentMethodMetadataTx(db *sql.Tx, pm *Method, createdBy string) error {
	if pm.ID == 0 {
		return ErrPaymentMethodWithoutID
	}
	m := metadata.MetadataFromValues(pm.Metadata, createdBy)
	return metadata.InsertMetadataTx(db, MetadataModel, pm.ID, m)
}

func PaymentMethodMetadataTx(db *sql.Tx, pm *Method) (map[string]string, error) {
	if pm.ID == 0 {
		return nil, ErrPaymentMethodWithoutID
	}
	m, err := metadata.MetadataByPrimaryTx(db, MetadataModel, pm.ID)
	if err != nil {
		return nil, err
	}
	return m.Values(), nil
}
