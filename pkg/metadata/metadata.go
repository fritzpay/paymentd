package metadata

import (
	"database/sql"
	"fmt"
	"time"
)

type MetadataModel interface {
	Schema() string
	PrimaryField() string
}
type MetadataEntry struct {
	Name      string
	Timestamp time.Time
	CreatedBy string
	Value     string
}
type Metadata map[string]MetadataEntry

const metadataByPrimary = `
SELECT
	md.name,
	md.timestamp,
	md.created_by,
	md.value
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

func MetadataByPrimaryDB(db *sql.DB, m MetadataModel, primary int64) (Metadata, error) {
	query := fmt.Sprintf(metadataByPrimary, m.Schema(), m.PrimaryField())
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
		e.Timestamp = time.Unix(0, t)
		metadata[e.Name] = e
	}
	err = rows.Err()
	rows.Close()
	return metadata, err
}

const insertMetadata = `
INSERT INTO %s
(%s, name, timestamp, created_by, value)
VALUES
(?, ?, ?, ?, ?)
`

func InsertMetadataTx(db *sql.Tx, m MetadataModel, primary int64, metadata Metadata) error {
	insert, err := db.Prepare(fmt.Sprintf(insertMetadata, m.Schema(), m.PrimaryField()))
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
