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
func InsertProjectDB(db *sql.DB, p *Project) error {
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
	principal_id,
	name,
	created,
	created_by
FROM project
`

const selectProjectById = selectProject + `
WHERE
	id = ?
`

const selectProjectByPrincipalIdAndName = selectProject + `
WHERE
	principal_id = ?
AND
	name = ?	
`

const selectProjectByName = selectProject + `
WHERE
	name = ?
`

func scanProject(row *sql.Row) (*Project, error) {
	p := &Project{}
	err := row.Scan(&p.ID, &p.PrincipalID, &p.Name, &p.Created, &p.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrProjectNotFound
		}
		return p, err
	}
	return p, nil
}

// ProjectByIdDB selects a project by the given project id
//
// If no such project exists, it will return an empty project
func ProjectByIdDB(db *sql.DB, projectId int64) (*Project, error) {
	row := db.QueryRow(selectProjectById, projectId)
	return scanProject(row)
}

// ProjectByIdTx selects a project by the given project id
//
// If no such project exists, it will return an empty project
func ProjectByIdTx(db *sql.Tx, projectId int64) (*Project, error) {
	row := db.QueryRow(selectProjectById, projectId)
	return scanProject(row)
}

// ProjectByName selects a project by the given project name
//
// If no such project exists, it will return an empty project
func ProjectByNameDB(db *sql.DB, projectName string) (*Project, error) {
	row := db.QueryRow(selectProjectByName, projectName)
	return scanProject(row)
}

// ProjectByNameTx selects a project by the given project name
//
// If no such project exists, it will return an empty project
func ProjectByNameTx(db *sql.Tx, projectName string) (*Project, error) {
	row := db.QueryRow(selectProjectByName, projectName)
	return scanProject(row)
}

const selectProjectKey = `
SELECT
	k.key,
	k.timestamp,
	k.created_by,
	k.secret,
	k.active,
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

func scanProjectKey(row *sql.Row) (*Projectkey, error) {
	pk := &Projectkey{}
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
func ProjectKeyByKeyDB(db *sql.DB, key string) (*Projectkey, error) {
	row := db.QueryRow(selectProjectKeyByKey, key)
	return scanProjectKey(row)
}
