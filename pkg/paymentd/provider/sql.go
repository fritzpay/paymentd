package provider

import (
	"database/sql"
	"errors"
)

var (
	ErrProviderNotFound = errors.New("provider not found")
)

const selectProvider = `
SELECT
	name
FROM provider
`

const selectProviderByName = selectProvider + `
WHERE
	name = ?
`

func scanSingleRow(row *sql.Row) (Provider, error) {
	p := Provider{}
	err := row.Scan(&p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrProviderNotFound
		}
		return p, err
	}
	return p, nil
}

func ProviderAllDB(db *sql.DB) ([]Provider, error) {
	rows, err := db.Query(selectProvider)

	d := make([]Provider, 0, 200)

	for rows.Next() {
		pr := Provider{}
		err := rows.Scan(&pr.Name)
		if err != nil {
			return d, err
		}
		d = append(d, pr)
	}

	return d, err
}

func ProviderByNameDB(db *sql.DB, name string) (Provider, error) {
	row := db.QueryRow(selectProviderByName, name)
	return scanSingleRow(row)
}

func ProviderByNameTx(db *sql.Tx, name string) (Provider, error) {
	row := db.QueryRow(selectProviderByName, name)
	return scanSingleRow(row)
}
