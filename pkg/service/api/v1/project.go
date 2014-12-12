package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fritzpay/paymentd/pkg/metadata"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

type ProjectAdminAPIResponse struct {
	AdminAPIResponse
}

// return a handler to add and manipulate projects
func (a *AdminAPI) ProjectRequest() http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log := a.log.New(log15.Ctx{"method": "ProjectRequest"})

		// @todo restrict by projectid
		switch r.Method {
		case "PUT":
			a.putNewProject(w, r)
		case "POST":
			a.postChangeProject(w, r)
		case "GET":
			a.getAllProjects(w, r)
		default:
			if Debug {
				log.Debug("request method not supported", log15.Ctx{"requestMethod": r.Method})
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return a.ctx.RateLimitHandler(h)
}

// return a hanlder to get project items
func (a *AdminAPI) ProjectGetRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log := a.log.New(log15.Ctx{"method": "ProjectGetRequest"})

		// @todo restrict by projectid
		if r.Method != "GET" {
			if Debug {
				log.Debug("request method not supported", log15.Ctx{"requestMethod": r.Method})
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		a.getProject(w, r)
	})
}

func (a *AdminAPI) getProject(w http.ResponseWriter, r *http.Request) {
	log := a.log.New(log15.Ctx{"method": "getProject"})

	// parse request paramter
	// project_id
	vars := mux.Vars(r)
	projectIDParam := vars["projectid"]
	projectID, err := strconv.ParseInt(projectIDParam, 10, 64)
	if err != nil {
		log.Error("param projectid conversion error", log15.Ctx{"err": err})
		ErrReadParam.Write(w)
		return
	}

	// get project from database
	db := a.ctx.PrincipalDB(service.ReadOnly)
	pr, err := project.ProjectByIDDB(db, projectID)
	if err == project.ErrProjectNotFound {
		log.Warn("project not found", log15.Ctx{"err": err})
		ErrNotFound.Write(w)
		return
	} else if err != nil {
		log.Error("get project from DB failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	md, err := metadata.MetadataByPrimaryDB(db, project.MetadataModel, pr.ID)
	if len(md) > 0 {
		pr.Metadata = md.Values()
	}
	if err != nil {
		log.Error("error retrieving metadata", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	pr.Metadata = md.Values()

	// response
	resp := ProjectAdminAPIResponse{}
	resp.Status = StatusSuccess
	resp.Info = "project " + pr.Name + " found"
	resp.Response = pr
	resp.Write(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("write error", log15.Ctx{"err": err})
		return
	}
}

func (a *AdminAPI) getAllProjects(w http.ResponseWriter, r *http.Request) {
	log := a.log.New(log15.Ctx{"method": "getAllProjects"})

	// parse request paramter
	// principal_id

	params := r.URL.Query()
	principalIDParam := params.Get("principalid")

	principalID, err := strconv.ParseInt(principalIDParam, 10, 64)
	if err != nil {
		log.Error("param principalid conversion error", log15.Ctx{"err": err})
		ErrReadParam.Write(w)
		return
	}

	// get projects from database
	db := a.ctx.PrincipalDB(service.ReadOnly)
	pr, err := project.AllProjectsByPrincipalIDDB(db, principalID)
	if err == project.ErrProjectNotFound {
		log.Warn("projects not found", log15.Ctx{"err": err})
		ErrNotFound.Write(w)
		return
	} else if err != nil {
		log.Error("get projects from DB failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	// Metadata required in general project listing?

	/*md, err := metadata.MetadataByPrimaryDB(db, project.MetadataModel, pr.ID)
	if len(md) > 0 {
		pr.Metadata = md.Values()
	}
	if err != nil {
		log.Error("error retrieving metadata", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	pr.Metadata = md.Values()*/

	// response
	resp := ProjectAdminAPIResponse{}
	resp.Status = StatusSuccess
	resp.Info = "project found"
	resp.Response = pr
	resp.Write(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("write error", log15.Ctx{"err": err})
		return
	}
}

// add new project
func (a *AdminAPI) putNewProject(w http.ResponseWriter, r *http.Request) {

	log := a.log.New(log15.Ctx{"method": "putNewProject"})
	auth, err := getAuthContainer(r)
	if err != nil {
		log.Crit("auth container error", log15.Ctx{"err": err})
		ErrSystem.Write(w)
		return
	}
	// parse put paramter
	jd := json.NewDecoder(r.Body)
	pr := project.Project{}
	err = jd.Decode(&pr)
	if err != nil {
		log.Warn("project parsing failed", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// created
	pr.CreatedBy = auth[AuthUserIDKey].(string)
	pr.Created = time.Now().UTC().Round(time.Second)

	// validate fields
	if len(pr.Name) < 1 {
		log.Warn("project without name")
		ErrInval.Write(w)
		return
	}
	if pr.PrincipalID == 0 {
		log.Warn("project without principal ID")
		ErrInval.Write(w)
		return
	}

	log = log.New(log15.Ctx{"projectName": pr.Name, "principalID": pr.PrincipalID})

	var tx *sql.Tx
	var commit bool
	defer func() {
		if tx != nil && !commit {
			err = tx.Rollback()
			if err != nil {
				log.Crit("error on rollback", log15.Ctx{"err": err})
			}
		}
	}()
	tx, err = a.ctx.PrincipalDB().Begin()
	if err != nil {
		commit = true
		log.Crit("error on begin", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	// does principal exist
	_, err = principal.PrincipalByIDTx(tx, pr.PrincipalID)
	if err != nil {
		if err == principal.ErrPrincipalNotFound {
			log.Warn("principal not found")
			resp := ErrNotFound
			resp.Info = "principal not found"
			resp.Write(w)
			return
		}
		log.Error("error retrieving principal", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	//check if this project already exist
	_, err = project.ProjectByPrincipalIDandIDTx(tx, pr.PrincipalID, pr.ID)
	if err != nil && err != project.ErrProjectNotFound {
		log.Error("error retrieving project", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	if err != project.ErrProjectNotFound {
		// project already exists
		log.Warn("project already exists", log15.Ctx{"err": err})
		ErrConflict.Write(w)
		return
	}

	// insert project from database
	err = project.InsertProjectTx(tx, &pr)
	if err != nil {
		log.Error("project creation failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	if pr.Config.HasValues() {
		err = project.InsertProjectConfigTx(tx, &pr)
		if err != nil {
			log.Error("error saving project config", log15.Ctx{"err": err})
			ErrDatabase.Write(w)
			return
		}
	}
	if pr.Metadata != nil {
		meta := metadata.MetadataFromValues(pr.Metadata, pr.CreatedBy)
		err = metadata.InsertMetadataTx(tx, project.MetadataModel, pr.ID, meta)
		if err != nil {
			log.Error("error saving metadata", log15.Ctx{"err": err})
			ErrDatabase.Write(w)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Crit("error on commit", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	commit = true

	// output
	je := json.NewEncoder(w)
	err = je.Encode(&pr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("json encode failed.", log15.Ctx{"err": err})
		return
	}
}

// add change project data
func (a *AdminAPI) postChangeProject(w http.ResponseWriter, r *http.Request) {

	log := a.log.New(log15.Ctx{"method": "postChangeProject"})

	auth, err := getAuthContainer(r)
	if err != nil {
		log.Crit("error on auth container", log15.Ctx{"err": err})
		ErrSystem.Write(w)
		return
	}

	// get data from post variables
	jd := json.NewDecoder(r.Body)
	pr := &project.Project{}
	err = jd.Decode(pr)
	r.Body.Close()
	if err != nil {
		ErrReadJson.Write(w)
		log.Error("json decode failed", log15.Ctx{"err": err})
		return
	}

	// created
	pr.CreatedBy = auth[AuthUserIDKey].(string)
	pr.Created = time.Now().UTC().Round(time.Second)

	// Rollback handling
	var tx *sql.Tx
	var commit bool
	defer func() {
		if tx != nil && !commit {
			err = tx.Rollback()
			if err != nil {
				log.Crit("error on rollback", log15.Ctx{"err": err})
			}
		}
	}()

	// adding the data
	tx, err = a.ctx.PrincipalDB().Begin()
	if err != nil {
		commit = true
		ErrDatabase.Write(w)
		log.Error("error on begin", log15.Ctx{"err": err})
	}

	//does project exist
	_, err = project.ProjectByPrincipalIDandIDTx(tx, pr.PrincipalID, pr.ID)
	if err == project.ErrProjectNotFound {
		log.Error("error retrieving project", log15.Ctx{"err": err})
		ErrInval.Write(w)
		return
	}
	if err != nil {
		log.Warn("database error", log15.Ctx{"err": err})
		ErrConflict.Write(w)
		return
	}

	// insert Metadata
	md := metadata.MetadataFromValues(pr.Metadata, pr.CreatedBy)
	err = metadata.InsertMetadataTx(tx, project.MetadataModel, pr.ID, md)
	if err != nil {
		log.Error("metadata insert failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	// get stored and added metadata from db
	pr, err = project.ProjectByPrincipalIDandIDTx(tx, pr.PrincipalID, pr.ID)
	if err != nil {
		log.Error("get metadata failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	md, err = metadata.MetadataByPrimaryTx(tx, project.MetadataModel, pr.ID)
	if err != nil {
		log.Error("get metadata failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	pr.Metadata = md.Values()

	err = tx.Commit()
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("error on commit", log15.Ctx{"err": err})
		return
	}

	commit = true

	// create response
	resp := ProjectAdminAPIResponse{}
	resp.Info = "project " + pr.Name + " changed"
	resp.Status = StatusSuccess
	resp.Response = pr
	err = resp.Write(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("write error", log15.Ctx{"err": err})
		return
	}
}
