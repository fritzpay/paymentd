package metadata

import (
	"database/sql"
	"fmt"
	"time"
)

const metadataFields = `
SELECT
	m.name,
	m.timestamp,
	m.created_by,
	m.value
`

const metadataByPrimary = metadataFields + `
FROM %[1]s AS m
WHERE
	m.%[2]s = ?
	AND
	m.timestamp = (
		SELECT MAX(timestamp) FROM %[1]s
		WHERE
			%[2]s = m.%[2]s
			AND
			name = m.name
	)
`

func MetadataByPrimaryDB(db *sql.DB, m MetadataModeler, primary int64) (Metadata, error) {
	query := fmt.Sprintf(metadataByPrimary, m.Table(), m.PrimaryField())
	rows, err := db.Query(query, primary)
	if err != nil {
		return nil, err
	}
	metadata := Metadata(make(map[string]MetadataEntry))
	var t int64
	var e MetadataEntry
	for rows.Next() {
		err = rows.Scan(&e.Name, &t, &e.CreatedBy, &e.Value)
		if err != nil {
			rows.Close()
			return nil, err
		}
		e.timestamp = time.Unix(0, t)
		metadata[e.Name] = e
	}
	err = rows.Err()
	rows.Close()
	return metadata, err
}

const metadataByPrimaryAndName = metadataFields + `
FROM %[1]s AS m
WHERE
	m.%[2]s = ?
	AND
	m.name = ?
	m.Tableimestamp = (
		SELECT MAX(timestamp) FROM %[1]s
		WHERE
			%[2]s = m.%[2]s
			AND
			name = m.name
	)
`

// MetadataByPrimaryAndNameDB selects a specific metadata entry
//
// If no such entry with the name exists, it will return an empty entry
func MetadataByPrimaryAndNameDB(db *sql.DB, m MetadataModeler, primary int64, name string) (MetadataEntry, error) {
	query := fmt.Sprintf(metadataByPrimaryAndName, m.Table(), m.PrimaryField())
	row := db.QueryRow(query, primary, name)
	var e MetadataEntry
	var ts int64
	err := row.Scan(&e.Name, &ts, &e.CreatedBy, &e.Value)
	if err != nil {
		if err == sql.ErrNoRows {
			return e, nil
		}
		return e, err
	}
	e.timestamp = time.Unix(0, ts)
	return e, nil
}

const insertMetadata = `
INSERT INTO %s
(%s, name, timestamp, created_by, value)
VALUES
(?, ?, ?, ?, ?)
`

func InsertMetadataTx(db *sql.Tx, m MetadataModeler, primary int64, metadata Metadata) error {
	insert, err := db.Prepare(fmt.Sprintf(insertMetadata, m.Table(), m.PrimaryField()))
	if err != nil {
		return err
	}
	t := time.Now()
	for _, e := range metadata {
		_, err = insert.Exec(primary, e.Name, t.UnixNano(), e.CreatedBy, e.Value)
		if err != nil {
			insert.Close()
			return err
		}
	}
	insert.Close()
	return nil
}
