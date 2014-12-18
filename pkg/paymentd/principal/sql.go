package principal

import (
	"database/sql"
	"errors"
	"time"
)

var (
	// ErrPrincipalNotFound is an error which various select methods will return
	// if the requested principal was not found
	ErrPrincipalNotFound = errors.New("principal not found")
)

const insertPrincipal = `
INSERT INTO principal
(created, created_by, name)
VALUES
(?, ?, ?)
`

func execInsertPrincipal(insert *sql.Stmt, p *Principal) error {
	res, err := insert.Exec(p.Created, p.CreatedBy, p.Name)
	if err != nil {
		insert.Close()
		return err
	}
	p.ID, err = res.LastInsertId()
	insert.Close()
	return err
}

// InsertPrincipalTx inserts a principal
//
// This will modify the given principal, setting the ID field.
func InsertPrincipalTx(db *sql.Tx, p *Principal) error {
	insert, err := db.Prepare(insertPrincipal)
	if err != nil {
		return err
	}
	return execInsertPrincipal(insert, p)
}

const selectPrincipal = `
SELECT
	pr.id,
	pr.created,
	pr.created_by,
	pr.name,
	s.status
FROM principal AS pr
INNER JOIN principal_status AS s ON
	s.principal_id = pr.id
	AND
	s.timestamp = (
		SELECT MAX(timestamp) FROM principal_status
		WHERE
			principal_id = pr.id
	)
`

const selectPrincipalByID = selectPrincipal + `
WHERE
	s.status <> '` + PrincipalStatusDeleted + `'
	AND
	pr.id = ?
`

func scanPrincipal(row *sql.Row) (Principal, error) {
	p := Principal{}
	err := row.Scan(&p.ID, &p.Created, &p.CreatedBy, &p.Name, &p.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrPrincipalNotFound
		}
		return p, err
	}
	return p, nil
}

func PrincipalByIDTx(db *sql.Tx, id int64) (Principal, error) {
	row := db.QueryRow(selectPrincipalByID, id)
	return scanPrincipal(row)
}

const selectPrincipalByName = selectPrincipal + `
WHERE
	s.status <> '` + PrincipalStatusDeleted + `'
	AND
	pr.name = ?
`

// PrincipalByNameDB selects a principal by the given name
//
// If no such principal exists, it will return an empty principal
func PrincipalByNameDB(db *sql.DB, name string) (Principal, error) {
	row := db.QueryRow(selectPrincipalByName, name)
	return scanPrincipal(row)
}

// PrincipalByNameTx selects a principal by the given name
//
// If no such principal exists, it will return an empty principal
func PrincipalByNameTx(db *sql.Tx, name string) (Principal, error) {
	row := db.QueryRow(selectPrincipalByName, name)
	return scanPrincipal(row)
}

const selectPrincipalIDByName = `
SELECT id FROM principal WHERE name = ?
`

func PrincipalIDByNameTx(db *sql.Tx, name string) (int64, error) {
	row := db.QueryRow(selectPrincipalIDByName, name)
	var id int64
	err := row.Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrPrincipalNotFound
		}
		return 0, err
	}
	return id, nil
}

const insertPrincipalStatus = `
INSERT INTO principal_status
(principal_id, timestamp, created_by, status)
VALUES
(?, ?, ?, ?)
`

// InsertPrincipalStatusTx adds a status entry for the given principal
func InsertPrincipalStatusTx(db *sql.Tx, pr Principal, createdBy string) error {
	stmt, err := db.Prepare(insertPrincipalStatus)
	if err != nil {
		return err
	}
	ts := time.Now()
	_, err = stmt.Exec(pr.ID, ts.UnixNano(), createdBy, pr.Status)
	stmt.Close()
	return err
}
