package paypal_rest

import (
	"database/sql"
	"errors"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
)

var (
	ErrConfigNotFound      = errors.New("config not found")
	ErrTransactionNotFound = errors.New("transaction not found")
)

const selectConfig = `
SELECT
	c.project_id,
	c.method_key,
	c.created,
	c.created_by,
	c.endpoint,
	c.client_id,
	c.secret,
	c.type
FROM provider_paypal_config AS c
`
const selectConfigByProjectIDAndMethodKey = selectConfig + `
WHERE
	c.project_id = ?
	AND
	c.method_key = ?
	AND
	c.created = (
		SELECT MAX(created) FROM provider_paypal_config
		WHERE
			project_id = c.project_id
			AND
			method_key = c.method_key
	)
`

func scanConfig(row *sql.Row) (*Config, error) {
	cfg := &Config{}
	err := row.Scan(
		&cfg.ProjectID,
		&cfg.MethodKey,
		&cfg.Created,
		&cfg.CreatedBy,
		&cfg.Endpoint,
		&cfg.ClientID,
		&cfg.Secret,
		&cfg.Type,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return cfg, ErrConfigNotFound
		}
		return cfg, err
	}
	return cfg, nil
}

func ConfigByPaymentMethodTx(db *sql.Tx, method *payment_method.Method) (*Config, error) {
	row := db.QueryRow(selectConfigByProjectIDAndMethodKey, method.ProjectID, method.MethodKey)
	return scanConfig(row)
}

func ConfigByPaymentMethodDB(db *sql.DB, method *payment_method.Method) (*Config, error) {
	row := db.QueryRow(selectConfigByProjectIDAndMethodKey, method.ProjectID, method.MethodKey)
	return scanConfig(row)
}

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

const selectTransactionByPaymentIDAndType = selectTransaction + `
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
			AND
			type = ?
	)
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

func TransactionByPaymentIDAndTypeTx(db *sql.Tx, paymentID payment.PaymentID, t string) (*Transaction, error) {
	row := db.QueryRow(selectTransactionByPaymentIDAndType, paymentID.ProjectID, paymentID.PaymentID, t)
	return scanTransactionRow(row)
}

func TransactionByPaymentIDAndTypeDB(db *sql.DB, paymentID payment.PaymentID, t string) (*Transaction, error) {
	row := db.QueryRow(selectTransactionByPaymentIDAndType, paymentID.ProjectID, paymentID.PaymentID, t)
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

const insertAuthorization = `
INSERT INTO paypal_authorization
(project_id, payment_id, timestamp, valid_until, state, authorization_id, paypal_id, amount, currency, links, data)
VALUES
(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

func InsertAuthorizationTx(db *sql.Tx, auth *Authorization) error {
	stmt, err := db.Prepare(insertAuthorization)
	if err != nil {
		return err
	}
	_, err = stmt.Exec(
		auth.ProjectID,
		auth.PaymentID,
		auth.Timestamp.UnixNano(),
		auth.ValidUntil,
		auth.State,
		auth.AuthorizationID,
		auth.PaypalID,
		auth.Amount,
		auth.Currency,
		auth.Links,
		auth.Data,
	)
	stmt.Close()
	return err
}
