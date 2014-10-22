package currency

import (
	"database/sql"
	"errors"
)

var (
	// ErrCurrencyNotFound is an error which various select methods will return
	// if the requested currency was not found
	ErrCurrencyNotFound = errors.New("currency not found")
)

const selectCurrency = `
SELECT
	code_iso_4217
FROM currency
`

const selectCurrencyByCodeIso4217 = selectCurrency + `
WHERE
	code_iso_4217 = ?
`

func scanCurrency(row *sql.Row) (Currency, error) {
	c := Currency{}
	err := row.Scan(&c.CodeISO4217)
	if err != nil {
		if err == sql.ErrNoRows {
			return c, ErrCurrencyNotFound
		}
		return c, err
	}
	return c, nil
}

// CurrencyByCodeIso4217DB selects a currency by the given code
//
// If no such currency exists, it will return an empty currency
func CurrencyByCodeISO4217DB(db *sql.DB, codeISO4217 string) (Currency, error) {
	row := db.QueryRow(selectCurrencyByCodeIso4217, codeISO4217)
	return scanCurrency(row)
}

// CurrencyByCodeIso4217Tx selects a currency by the given code
//
// If no such currency exists, it will return an empty currency
func CurrencyByCodeISO4217Tx(db *sql.Tx, codeISO4217 string) (Currency, error) {
	row := db.QueryRow(selectCurrencyByCodeIso4217, codeISO4217)
	return scanCurrency(row)
}

// CurrencyAllDB selects all available currencies
func CurrencyAllDB(db *sql.DB) ([]Currency, error) {
	rows, err := db.Query(selectCurrency)
	d := make([]Currency, 0, 200)
	for rows.Next() {
		c := Currency{}
		err := rows.Scan(&c.CodeISO4217)
		if err != nil {
			return d, err
		}
		d = append(d, c)
	}
	return d, err
}
