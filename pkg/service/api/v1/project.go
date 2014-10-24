package v1

import (
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"path"
	"strconv"
)

// return a handler brokering the project related admin api requests
func (a *AdminAPI) ProjectRequest() http.Handler {

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log := a.log.New(log15.Ctx{"method": "Project Request"})
		log.Info("Method:" + r.Method)

		// @todo ristrict by projectid
		if r.Method == "GET" {
			a.getProject(w, r)
		} else if r.Method == "PUT" {
			a.putNewProject(w, r)
		} else if r.Method == "POST" {
			a.postChangeProject(w, r)
		} else {
			log.Info("request method not supported: " + r.Method)
			w.WriteHeader(http.StatusBadRequest)
		}

	})

	return h
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
		log.Error("project id convertion failed", log15.Ctx{"err": err})
		// @todo if param is not numeric try project_name
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get project from database
	db := a.ctx.PrincipalDB(service.ReadOnly)
	pr, err := project.ProjectByIdDB(db, projectId)

	if err == project.ErrProjectNotFound {
		log.Error("project not found", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		log.Error("get project from DB failed", log15.Ctx{"err": err})
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
	log.Info("post project")

}
