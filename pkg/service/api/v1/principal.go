package v1

import (
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/metadata"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"path"
	"time"
)

type PrincipalAdminAPIResponse struct {
	AdminAPIResponse
}

// handler to create or change a principal
//
// PUT creates new principal
// POST can be used to change the principals metadata
func (a *AdminAPI) PrincipalRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{"method": "Principal Request"})
		if r.Method == "PUT" {
			a.putNewPrincipal(w, r)
		} else if r.Method == "POST" {
			a.postChangePrincipal(w, r)
		} else {
			log.Info("http method not supported: " + r.Method)
			ErrMethod.Write(w)
		}
	})
}

// handler to display a specific existing principal
func (a *AdminAPI) PrincipalGetRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{"method": "principal GET request"})
		log.Info("Method:" + r.Method)
		// get principal by name
		urlpath, principalName := path.Split(path.Clean(r.URL.Path))

		log.Info("principalName: " + principalName)
		log.Info("url path: " + urlpath)

		db := a.ctx.PrincipalDB(service.ReadOnly)
		pr, err := principal.PrincipalByNameDB(db, principalName)
		if err == principal.ErrPrincipalNotFound {
			ErrNotFound.Write(w)
			log.Info("not found.", log15.Ctx{"err": err})
			return
		}
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("DB get by name failed.", log15.Ctx{"err": err})
			return
		}
		md, err := metadata.MetadataByPrimaryDB(db, principal.MetadataModel, pr.ID)
		if err != nil {
			ErrNotFound.Write(w)
			log.Error("get metadata failed.", log15.Ctx{"err": err})
			return
		}
		if len(md) > 0 {
			pr.Metadata = md.Values()
		}

		// create service response object
		resp := PrincipalAdminAPIResponse{}
		resp.Status = StatusSuccess
		resp.Info = "principal " + pr.Name + " found"
		resp.Response = pr
		err = resp.Write(w)
		if err != nil {
			log.Error("write error", log15.Ctx{"err": err})

			return
		}
	})
}

func (a *AdminAPI) putNewPrincipal(w http.ResponseWriter, r *http.Request) {
	log := a.log.New(log15.Ctx{"method": "Principal Request"})
	log.Info("Method:" + r.Method)

	// create new principal
	jd := json.NewDecoder(r.Body)
	pr := principal.Principal{}
	err := jd.Decode(&pr)
	r.Body.Close()
	if err != nil {
		ErrReadJson.Write(w)
		log.Error("json decode failed", log15.Ctx{"err": err})
		return
	}

	// validation createdBy has to be set
	if len(pr.CreatedBy) < 1 {
		ErrInval.Info = "CreatedBy has to be set"
		log.Info("CreatedBy has to be set:" + pr.Name)
		return
	}
	// set created time
	pr.Created = time.Now()

	// check if principal exists
	db := a.ctx.PrincipalDB()
	_, err = principal.PrincipalByNameDB(db, pr.Name)
	if err == principal.ErrPrincipalNotFound {
		// insert pr if not exists
		// start dbtx to save metadata and pr together
		tx, err := db.Begin()
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("Tx begin failed.", log15.Ctx{"err": err})
		}

		// insert pr
		err = principal.InsertPrincipalTx(tx, &pr)
		if err != nil {
			tx.Rollback()
			ErrSystem.Write(w)
			log.Error("TX insert failed.", log15.Ctx{"err": err})
			return
		}

		// insert metadata
		md := metadata.MetadataFromValues(pr.Metadata, pr.CreatedBy)
		err = metadata.InsertMetadataTx(tx, principal.MetadataModel, pr.ID, md)
		if err != nil {
			tx.Rollback()
			ErrDatabase.Write(w)
			log.Error("TX metadata insert failed.", log15.Ctx{"err": err})
			return
		}
		//commit tx
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			ErrSystem.Write(w)
			log.Error("TX commit failed.", log15.Ctx{"err": err})
			return
		}
	} else if err != nil {
		// other db error
		ErrSystem.Write(w)
		log.Error("DB get by name failed.", log15.Ctx{"err": err})
		return
	} else {
		// already exists
		ErrConflict.Write(w)
		log.Warn("principal already exist: " + string(pr.ID) + " " + pr.Name)
		return
	}

	// get data explicit from DB
	pr, err = principal.PrincipalByNameDB(db, pr.Name)
	if err != nil {
		ErrSystem.Write(w)
		log.Error("DB get by name failed.", log15.Ctx{"err": err})
		return
	}
	md, err := metadata.MetadataByPrimaryDB(db, principal.MetadataModel, pr.ID)
	if len(md) > 0 {
		pr.Metadata = md.Values()
	}
	if err != nil {
		log.Error("get metadata failed.", log15.Ctx{"err": err})
		return
	}

	resp := PrincipalAdminAPIResponse{}
	resp.HttpStatus = http.StatusOK
	resp.Status = StatusSuccess
	resp.Response = pr
	resp.Info = "principal " + pr.Name + " created"
	err = resp.Write(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("write error", log15.Ctx{"err": err})
		return
	}
}

// post method to add and change the metadata
func (a *AdminAPI) postChangePrincipal(w http.ResponseWriter, r *http.Request) {
	log := a.log.New(log15.Ctx{"method": "Principal Request"})
	log.Info("Method:" + r.Method)

	// get Metadata from post variables
	jd := json.NewDecoder(r.Body)
	pr := principal.Principal{}
	err := jd.Decode(&pr)
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

	// does principal exist
	db := a.ctx.PrincipalDB(service.ReadOnly)
	prdb, err := principal.PrincipalByNameDB(db, pr.Name)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Error("get principal from DB failed:"+pr.Name, log15.Ctx{"err": err})
		return
	}
	pr.ID = prdb.ID

	// open transaction to add the posted metadata
	tx, err := db.Begin()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("Tx begin failed.", log15.Ctx{"err": err})
	}
	md := metadata.MetadataFromValues(postedMetadata, pr.CreatedBy)

	err = metadata.InsertMetadataTx(tx, principal.MetadataModel, pr.ID, md)
	if err != nil {
		tx.Rollback()
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("insert metadata failed", log15.Ctx{"err": err})
		return
	}
	tx.Commit()

	// get stored data from db
	pr, err = principal.PrincipalByNameDB(db, pr.Name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		log.Error("DB get by name failed.", log15.Ctx{"err": err})
		return
	}
	md, err = metadata.MetadataByPrimaryDB(db, principal.MetadataModel, pr.ID)
	if len(md) > 0 {
		pr.Metadata = md.Values()
	}
	if err != nil {
		log.Error("get metadata failed.", log15.Ctx{"err": err})
		return
	}

	// create response
	resp := PrincipalAdminAPIResponse{}
	resp.Info = "principal " + pr.Name + " changed"
	resp.Status = StatusSuccess
	resp.Response = pr
	err = resp.Write(w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("write error", log15.Ctx{"err": err})
		return
	}
}
