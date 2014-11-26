package paypal_rest

import (
	"database/sql"
	"errors"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
)

var (
	ErrConfigNotFound = errors.New("config not found")
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
