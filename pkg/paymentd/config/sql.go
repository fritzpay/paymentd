package config

import (
	"database/sql"
	"time"
)

const selectEntryCountByName = `
SELECT COUNT(*) FROM config 
WHERE 
	name = ?
	AND
	last_change = (
		SELECT MAX(last_change) FROM config AS mc
		WHERE mc.name = name
	)
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

func readSingleEntry(row *sql.Row) (Entry, error) {
	e := Entry{}
	var ts int64
	err := row.Scan(&e.Name, &ts, &e.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return e, ErrEntryNotFound
		}
		return e, err
	}
	e.lastChange = time.Unix(0, ts)
	return e, nil
}

// EntryByNameDB selects the current configuration entry for the given name
// If the name is not present, it returns nil
func EntryByNameDB(db *sql.DB, name string) (Entry, error) {
	row := db.QueryRow(selectEntryByName, name)
	return readSingleEntry(row)
}

// EntryByNameTx selects the current configuration entry for the given name
// If the name is not present, it returns nil
//
// This function should be used inside a (SQL-)transaction
func EntryByNameTx(db *sql.Tx, name string) (Entry, error) {
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
	_, err = stmt.Exec(e.Name, t.UnixNano(), e.Value)
	stmt.Close()
	return err
}

// InsertConfigTx saves a config set
//
// This function should be used inside a (SQL-)transaction
func InsertConfigTx(db *sql.Tx, cfg Config) error {
	stmt, err := db.Prepare(insertEntry)
	if err != nil {
		return err
	}
	t := time.Now()
	for n, v := range cfg {
		_, err = stmt.Exec(n, t.UnixNano(), v)
		if err != nil {
			stmt.Close()
			return err
		}
	}
	stmt.Close()
	return nil
}

// InsertConfigIfNotPresentTx saves a config set if the names are not present
//
// This funtion should be used inside a (SQL-)transaction
func InsertConfigIfNotPresentTx(db *sql.Tx, cfg Config) error {
	checkExists, err := db.Prepare(selectEntryCountByName)
	if err != nil {
		return err
	}
	defer checkExists.Close()
	insert, err := db.Prepare(insertEntry)
	if err != nil {
		return err
	}
	defer insert.Close()
	var exists *sql.Row
	var numEntries int
	t := time.Now()
	for n, v := range cfg {
		exists = checkExists.QueryRow(n)
		err = exists.Scan(&numEntries)
		if err != nil {
			return err
		}
		if numEntries > 0 {
			continue
		}
		_, err = insert.Exec(n, t.UnixNano(), v)
		if err != nil {
			return err
		}
	}
	return nil
}
