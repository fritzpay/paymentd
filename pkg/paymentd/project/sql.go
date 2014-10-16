package project

import (
	"database/sql"
)

const insertProject = `
INSERT INTO project
(principal_id, created, created_by, name)
VALUES
(?, ?, ?,?)
`

func execInsertProject(insert *sql.Stmt, p *Project) error {
	res, err := insert.Exec(p.PrincipalID, p.Created, p.CreatedBy, p.Name)
	if err != nil {
		insert.Close()
		return err
	}
	p.ID, err = res.LastInsertId()
	insert.Close()
	return err
}

// InsertProjectDB inserts a project
//
// This will modify the given project, setting the ID field.
func InsertPrincipalDB(db *sql.DB, p *Project) error {
	insert, err := db.Prepare(insertProject)
	if err != nil {
		return err
	}
	return execInsertProject(insert, p)
}

// InsertProjectTx inserts a project
//
// This will modify the given project, setting the ID field.
func InsertProjectTx(db *sql.Tx, p *Project) error {
	insert, err := db.Prepare(insertProject)
	if err != nil {
		return err
	}
	return execInsertProject(insert, p)
}

const selectProject = `
SELECT
	id,
	created,
	created_by,
	name
FROM project
`

const selectProjectByPrincipalIDAndName = selectProject + `
WHERE
	principal_id = ?
AND
	name = ?
`

func scanProject(row *sql.Row) (Project, error) {
	p := Project{}
	err := row.Scan(&p.ID, &p.PrincipalID, &p.Created, &p.CreatedBy, &p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, nil
		}
		return p, err
	}
	return p, nil
}

// ProjectByPrincipalAndNameDB selects a project by the given
// principal id and project name
//
// If no such project exists, it will return an empty project
func ProjectByPrincipalIdAndNameDB(db *sql.DB, principalID int64, name string) (Project, error) {
	row := db.QueryRow(selectProjectByPrincipalIDAndName, principalID, name)
	return scanProject(row)
}

// ProjectByPrincipalAndNameDB selects a project by the given
// principal id and project name
//
// If no such principal exists, it will return an empty project
func ProjectByPrincipalIdAndNameTx(db *sql.Tx, principalID int64, name string) (Project, error) {
	row := db.QueryRow(selectProjectByPrincipalIDAndName, name)
	return scanProject(row)
}
