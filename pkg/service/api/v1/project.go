package v1

import (
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/metadata"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
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

		log := a.log.New(log15.Ctx{"method": "Project Request"})
		log.Info("Method:" + r.Method)

		// @todo restrict by projectid
		if r.Method == "PUT" {
			a.putNewProject(w, r)
		} else if r.Method == "POST" {
			a.postChangeProject(w, r)
		} else {
			log.Info("request method not supported: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}

	})

	return h
}

// return a hanlder to get project items
func (a *AdminAPI) ProjectGetRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log := a.log.New(log15.Ctx{"method": "Project Request"})
		log.Info("Method:" + r.Method)

		// @todo restrict by projectid
		if r.Method == "GET" {
			a.getProject(w, r)
		} else {
			log.Info("request method not supported: " + r.Method)
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func (a *AdminAPI) getProject(w http.ResponseWriter, r *http.Request) {

	log := a.log.New(log15.Ctx{"method": "Project request GET"})

	// parse request paramter
	// project_id
	urlpath, projectIdParam := path.Split(path.Clean(r.URL.Path))
	log.Info("path: " + urlpath)
	log.Info("project id: " + projectIdParam)
	projectId, err := strconv.ParseInt(projectIdParam, 10, 64)
	if err != nil {
		ErrReadParam.Write(w)
		log.Error("param conversion error", log15.Ctx{"err": err})
		return
	}

	// get project from database
	db := a.ctx.PrincipalDB(service.ReadOnly)
	pr, err := project.ProjectByIdDB(db, projectId)
	if err == project.ErrProjectNotFound {
		log.Error("project not found", log15.Ctx{"err": err})
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
		log.Error("metadata problem data not found", log15.Ctx{"err": err})
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

	log := a.log.New(log15.Ctx{"method": "Project request PUT"})

	// parse put paramter
	jd := json.NewDecoder(r.Body)
	pr := project.Project{}
	err := jd.Decode(&pr)
	if err != nil {
		log.Error("project parsing failed", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// validate fields
	if !pr.IsValid() {

		log.Error("project values not valid: Name:" + pr.Name + " CreatedBy:" + pr.CreatedBy)
		w.WriteHeader(http.StatusBadRequest)
		return

	}

	// get project from database
	db := a.ctx.PrincipalDB()

	//check if this project already exist
	_, err = project.ProjectByNameDB(db, pr.Name)
	if err == project.ErrProjectNotFound {
		// insert project from database
		err = project.InsertProjectDB(db, &pr)
		if err != nil {
			log.Error("project creation failed", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// output
		je := json.NewEncoder(w)
		err = je.Encode(&pr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("json encode failed.", log15.Ctx{"err": err})
			return
		}
	} else {
		// project already exists
		w.WriteHeader(http.StatusConflict)
		log.Error("project: "+string(pr.ID)+" "+pr.Name+" already exists.", log15.Ctx{"err": err})
		return
	}
}

func (a *AdminAPI) postChangeProject(w http.ResponseWriter, r *http.Request) {
	log := a.log.New(log15.Ctx{"method": "Project request POST"})
	log.Info("Method:" + r.Method)

	// get Metadata from post variables
	jd := json.NewDecoder(r.Body)
	pr := &project.Project{}
	err := jd.Decode(pr)
	r.Body.Close()
	if err != nil {
		ErrReadJson.Write(w)
		log.Error("json decode failed: ", log15.Ctx{"err": err})
		return
	}
	postedMetadata := pr.Metadata

	// validation createdBy has to be set
	if len(pr.CreatedBy) < 1 {
		ErrInval.Write(w)
		log.Info("CreatedBy has to be set:" + pr.Name)
		return
	}

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

func (a *AdminAPI) PaymentMethodsGetRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log := a.log.New(log15.Ctx{"method": "Project payment methods GET"})

		vars := mux.Vars(r)
		projectIdParam := vars["projectid"]
		projectId, err := strconv.ParseInt(projectIdParam, 10, 64)
		if err != nil {
			ErrReadParam.Write(w)
			log.Error("param conversion error", log15.Ctx{"err": err})
			return
		}

		// does project exist
		db := a.ctx.PrincipalDB(service.ReadOnly)
		var prdb *project.Project
		prdb, err = project.ProjectByIdDB(db, projectId)
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("database request failed", log15.Ctx{"err": err})
			return
		}

		// return methods
		resp := ProjectAdminAPIResponse{}
		resp.HttpStatus = http.StatusOK
		resp.Info = "project found"
		resp.Response = prdb
		resp.Write(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("write error", log15.Ctx{"err": err})
			return
		}
	})
}
func (a *AdminAPI) PaymentMethodsRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		log := a.log.New(log15.Ctx{"method": "Project payment method request"})

		// PUT create new entry
		// get parameters
		// projectid and methodname
		if r.Method == "PUT" {

			// check parameters exits in db
			// save values and metadata
			vars := mux.Vars(r)
			projectIdParam := vars["projectid"]

			projectId, err := strconv.ParseInt(projectIdParam, 10, 64)
			if err != nil {
				ErrReadParam.Write(w)
				log.Info("param not readable" + projectIdParam)
				return
			}
			db := a.ctx.PrincipalDB()
			pr, err := project.ProjectByIdDB(db, projectId)
			if err != nil {

			}

			resp := ProjectAdminAPIResponse{}
			resp.Status = StatusSuccess
			resp.Info = ""
			resp.Response = pr

		} else if r.Method == "POST" {

			// POST change
			// set status
			// set metadata
		} else {
			ErrMethod.Write(w)
			log.Error("http method not supported")
			return
		}

	})
}
