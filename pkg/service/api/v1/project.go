package v1

import (
	"database/sql"
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/metadata"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"path"
	"strconv"
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
		default:
			if Debug {
				log.Debug("request method not supported", log15.Ctx{"requestMethod": r.Method})
			}
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return h
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
	urlpath, projectIdParam := path.Split(path.Clean(r.URL.Path))
	if Debug {
		log.Debug("request", log15.Ctx{"requestPath": urlpath, "projectID": projectIdParam})
	}
	projectId, err := strconv.ParseInt(projectIdParam, 10, 64)
	if err != nil {
		log.Error("param conversion error", log15.Ctx{"err": err})
		ErrReadParam.Write(w)
		return
	}

	// get project from database
	db := a.ctx.PrincipalDB(service.ReadOnly)
	pr, err := project.ProjectByIdDB(db, projectId)
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
	pr.CreatedBy = auth[AuthUserIDKey].(string)

	log = log.New(log15.Ctx{"projectName": pr.Name})

	// validate fields
	if !pr.IsValid() {
		log.Warn("project values not valid")
		ErrInval.Write(w)
		return
	}
	if pr.Config.HasValues() {
		err = pr.Config.Validate()
		if err != nil {
			log.Warn("config not acceptable", log15.Ctx{"err": err})
			resp := ErrInval
			resp.Info = "config not acceptable"
			resp.Write(w)
			return
		}
	}

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
	}

	//check if this project already exist
	_, err = project.ProjectByNameTx(tx, pr.Name)
	if err != nil && err != project.ErrProjectNotFound {
		log.Error("error retrieving project by name", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	if err != project.ErrProjectNotFound {
		// project already exists
		log.Warn("project already exists.", log15.Ctx{"err": err})
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

func (a *AdminAPI) postChangeProject(w http.ResponseWriter, r *http.Request) {
	log := a.log.New(log15.Ctx{"method": "Project request POST"})
	log.Info("Method:" + r.Method)
	auth := service.RequestContextAuth(r)

	// get Metadata from post variables
	jd := json.NewDecoder(r.Body)
	pr := &project.Project{}
	err := jd.Decode(pr)
	r.Body.Close()
	pr.CreatedBy = auth[AuthUserIDKey].(string)
	if err != nil {
		ErrReadJson.Write(w)
		log.Error("json decode failed: ", log15.Ctx{"err": err})
		return
	}
	postedMetadata := pr.Metadata

	// does project exist
	db := a.ctx.PrincipalDB(service.ReadOnly)
	var prdb *project.Project

	prdb, err = project.ProjectByNameDB(db, pr.Name)
	if err == project.ErrProjectNotFound {
		ErrInval.Write(w)
		log.Info("project does not exist: "+pr.Name, log15.Ctx{"err": err})
		return
	}
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("get project from DB failed: "+pr.Name, log15.Ctx{"err": err})
		return
	}
	pr.ID = prdb.ID

	// open transaction to add the posted metadata
	tx, err := db.Begin()
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("Tx begin failed.", log15.Ctx{"err": err})
	}
	md := metadata.MetadataFromValues(postedMetadata, pr.CreatedBy)

	err = metadata.InsertMetadataTx(tx, project.MetadataModel, pr.ID, md)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("insert metadata failed", log15.Ctx{"err": err})
		return
	}
	tx.Commit()

	// get stored data from db
	pr, err = project.ProjectByNameDB(db, pr.Name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Error("DB get by name failed.", log15.Ctx{"err": err})
		return
	}
	md, err = metadata.MetadataByPrimaryDB(db, project.MetadataModel, pr.ID)
	if len(md) > 0 {
		pr.Metadata = md.Values()
	}
	if err != nil {
		log.Error("get metadata failed.", log15.Ctx{"err": err})
		return
	}

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
