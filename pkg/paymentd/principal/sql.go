package principal

import (
	"database/sql"
	"errors"
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

// InsertPrincipalDB inserts a principal
//
// This will modify the given principal, setting the ID field.
func InsertPrincipalDB(db *sql.DB, p *Principal) error {
	insert, err := db.Prepare(insertPrincipal)
	if err != nil {
		return err
	}
	return execInsertPrincipal(insert, p)
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
	id,
	created,
	created_by,
	name
FROM principal
`

const selectPrincipalByID = selectPrincipal + `
WHERE
	id = ?
`

func scanPrincipal(row *sql.Row) (Principal, error) {
	p := Principal{}
	err := row.Scan(&p.ID, &p.Created, &p.CreatedBy, &p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrPrincipalNotFound
		}
		return p, err
	}
	return p, nil
}

func PrincipalAllDB(db *sql.DB) ([]Principal, error) {
	rows, err := db.Query(selectPrincipal)
	if err != nil {
		return nil, err
	}

	d := make([]Principal, 0, 50)

	for rows.Next() {

		p := Principal{}
		err := rows.Scan(&p.ID, &p.Created, &p.CreatedBy, &p.Name)
		if err != nil {
			rows.Close()
			return d, err
		}
		d = append(d, p)

	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return d, err
	}
	if len(d) < 1 {
		return nil, ErrPrincipalNotFound
	}
	rows.Close()

	return d, err
}

func PrincipalByIDTx(db *sql.Tx, id int64) (Principal, error) {
	row := db.QueryRow(selectPrincipalByID, id)
	return scanPrincipal(row)
}

func PrincipalByIDDB(db *sql.DB, id int64) (Principal, error) {
	row := db.QueryRow(selectPrincipalByID, id)
	return scanPrincipal(row)
}

const selectPrincipalByName = selectPrincipal + `
WHERE
	name = ?
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
