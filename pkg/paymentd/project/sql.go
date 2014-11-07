package project

import (
	"database/sql"
	"errors"
	"time"
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

const insertProjectConfig = `
INSERT INTO project_config
(project_id, timestamp, web_url, callback_url, callback_api_version, project_key, return_url)
VALUES
(?, ?, ?, ?, ?, ?, ?)
`

func execInsertProjectConfig(insert *sql.Stmt, p *Project) error {
	p.Config.Timestamp = time.Now()
	_, err := insert.Exec(
		p.ID,
		p.Config.Timestamp,
		p.Config.WebURL,
		p.Config.CallbackURL,
		p.Config.CallbackAPIVersion,
		p.Config.ProjectKey,
		p.Config.ReturnURL,
	)
	insert.Close()
	return err
}

// InsertProjectConfigDB sets a new project config
//
// It will update the project config timestamp
func InsertProjectConfigDB(db *sql.DB, p *Project) error {
	insert, err := db.Prepare(insertProjectConfig)
	if err != nil {
		return err
	}
	return execInsertProjectConfig(insert, p)
}

// InsertProjectConfigTx sets a new project config
//
// It will update the project config timestamp
func InsertProjectConfigTx(db *sql.Tx, p *Project) error {
	insert, err := db.Prepare(insertProjectConfig)
	if err != nil {
		return err
	}
	return execInsertProjectConfig(insert, p)
}

const selectProject = `
SELECT
	p.id,
	p.principal_id,
	p.name,
	p.created,
	p.created_by,
	UNIX_TIMESTAMP(c.timestamp),
	c.web_url,
	c.callback_url,
	c.callback_api_version,
	c.project_key,
	c.return_url
FROM project AS p
LEFT JOIN project_config AS c ON
	c.project_id = p.id
	AND
	c.timestamp = (
		SELECT MAX(timestamp) FROM project_config
		WHERE
			project_id = p.id
	)
`

const selectProjectByPrincipalIDAndId = selectProject + `
WHERE
	principal_id = ?
AND
	id = ?
`

const selectProjectByPrincipalIdAndName = selectProject + `
WHERE
	principal_id = ?
AND
	name = ?	
`

func scanProject(row *sql.Row) (*Project, error) {
	p := &Project{}
	var ts sql.NullInt64
	err := row.Scan(
		&p.ID,
		&p.PrincipalID,
		&p.Name,
		&p.Created,
		&p.CreatedBy,
		&ts,
		&p.Config.WebURL,
		&p.Config.CallbackURL,
		&p.Config.CallbackAPIVersion,
		&p.Config.ProjectKey,
		&p.Config.ReturnURL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return p, ErrProjectNotFound
		}
		return p, err
	}
	if ts.Valid {
		p.Config.Timestamp = time.Unix(ts.Int64, 0)
	}
	return p, nil
}

// ProjectByIdDB selects a project by the given project id
//
// If no such project exists, it will return an empty project
func ProjectByPrincipalIDandIDDB(db *sql.DB, principalID int64, projectId int64) (*Project, error) {
	row := db.QueryRow(selectProjectByPrincipalIDAndId, principalID, projectId)
	return scanProject(row)
}

// ProjectByIdTx selects a project by the given project id
//
// If no such project exists, it will return an empty project
func ProjectByPrincipalIDandIDTx(db *sql.Tx, principalID int64, projectId int64) (*Project, error) {
	row := db.QueryRow(selectProjectByPrincipalIDAndId, principalID, projectId)
	return scanProject(row)
}

// ProjectByName selects a project by the given project name
//
// If no such project exists, it will return an empty project
func ProjectByPrincipalIDNameDB(db *sql.DB, principalID int64, projectName string) (*Project, error) {
	row := db.QueryRow(selectProjectByPrincipalIdAndName, principalID, projectName)
	return scanProject(row)
}

// ProjectByNameTx selects a project by the given project name
//
// If no such project exists, it will return an empty project
func ProjectByPrincipalIDAndNameTx(db *sql.Tx, principalID int64, projectName string) (*Project, error) {
	row := db.QueryRow(selectProjectByPrincipalIdAndName, principalID, projectName)
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
	p.created_by,
	UNIX_TIMESTAMP(c.timestamp),
	c.web_url,
	c.callback_url,
	c.callback_api_version,
	c.project_key,
	c.return_url
FROM project_key AS k
INNER JOIN project AS p ON
	p.id = k.project_id
LEFT JOIN project_config AS c ON
	c.project_id = p.id
	AND
	c.timestamp = (
		SELECT MAX(timestamp) FROM project_config
		WHERE
			project_id = p.id
	)
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
	var ts sql.NullInt64
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
		&ts,
		&pk.Project.Config.WebURL,
		&pk.Project.Config.CallbackURL,
		&pk.Project.Config.CallbackAPIVersion,
		&pk.Project.Config.ProjectKey,
		&pk.Project.Config.ReturnURL,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return pk, ErrProjectKeyNotFound
		}
		return pk, err
	}
	if ts.Valid {
		pk.Project.Config.Timestamp = time.Unix(ts.Int64, 0)
	}
	return pk, nil
}

// ProjectKeyByKeyDB selects a project key by the given key
func ProjectKeyByKeyDB(db *sql.DB, key string) (*Projectkey, error) {
	row := db.QueryRow(selectProjectKeyByKey, key)
	return scanProjectKey(row)
}
