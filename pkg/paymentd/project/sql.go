package project

import (
	"database/sql"
	"errors"
)

var (
	// ErrProjectNotFound will be returned by select functions when the requested
	// project was not found
	ErrProjectNotFound = errors.New("project not found")
	// ErrProjectKeyNotFound will be returned by select functions when the requested
	// project key was not found
	ErrProjectKeyNotFound = errors.New("project key not found")
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
			return p, ErrProjectNotFound
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

const selectProjectKey = `
SELECT
	k.key,
	k.timestamp,
	k.created_by,
	k.secret,
	k.active
	p.id,
	p.principal_id,
	p.name,
	p.created,
	p.created_by
FROM project_key AS k
INNER JOIN project AS p ON
	p.id = k.project_id
`

const selectProjectKeyByKey = selectProjectKey + `
WHERE
	k.key = ?
	AND
	k.timestamp = (
		SELECT MAX(timestamp) FROM project_key AS mk
		WHERE
			mk.key = k.key
	)
`

func scanProjectKey(row *sql.Row) (Projectkey, error) {
	pk := Projectkey{}
	err := row.Scan(
		&pk.Key,
		&pk.Timestamp,
		&pk.CreatedBy,
		&pk.Secret,
		&pk.Active,
		&pk.Project.ID,
		&pk.Project.PrincipalID,
		&pk.Project.Name,
		&pk.Project.Created,
		&pk.Project.CreatedBy,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return pk, ErrProjectKeyNotFound
		}
		return pk, err
	}
	return pk, nil
}

// ProjectKeyByKeyDB selects a project key by the given key
func ProjectKeyByKeyDB(db *sql.DB, key string) (Projectkey, error) {
	row := db.QueryRow(selectProjectKeyByKey, key)
	return scanProjectKey(row)
}
