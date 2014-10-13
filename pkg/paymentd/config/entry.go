package config

import (
	"database/sql"
	"time"
)

// Entry represents a configuration entry
type Entry struct {
	Name       string
	lastChange time.Time
	Value      string
}

const selectEntryCountByName = `
SELECT COUNT(*) FROM config WHERE name = ?
`

const selectEntryByName = `
SELECT
	c.name,
	c.last_change,
	c.value
FROM config AS c
WHERE
	c.name = ?
	AND
	c.last_change = (
		SELECT MAX(last_change) FROM config AS mc
		WHERE mc.name = c.name
	)
`

func readSingleEntry(row *sql.Row) (*Entry, error) {
	e := &Entry{}
	err := row.Scan(&e.Name, &e.lastChange, &e.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return e, nil
}

// EntryByNameDB selects the current configuration entry for the given name
// If the name is not present, it returns nil
func EntryByNameDB(db *sql.DB, name string) (*Entry, error) {
	row := db.QueryRow(selectEntryByName, name)
	return readSingleEntry(row)
}

// EntryByNameTx selects the current configuration entry for the given name
// If the name is not present, it returns nil
//
// This function should be used inside a (SQL-)transaction
func EntryByNameTx(db *sql.Tx, name string) (*Entry, error) {
	row := db.QueryRow(selectEntryByName, name)
	return readSingleEntry(row)
}

const insertEntry = `
INSERT INTO config
(name, last_change, value)
VALUES
(?, ?, ?)
`

// InsertEntryDB inserts an entry
func InsertEntryDB(db *sql.DB, e Entry) error {
	stmt, err := db.Prepare(insertEntry)
	if err != nil {
		return err
	}
	t := time.Now()
	_, err = stmt.Exec(e.Name, t, e.Value)
	stmt.Close()
	return err
}
