package v1

import (
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/metadata"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
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
		log := a.log.New(log15.Ctx{"method": "PrincipalRequest"})

		switch r.Method {
		case "PUT":
			a.putNewPrincipal(w, r)
		case "POST":
			a.postChangePrincipal(w, r)
		default:
			ErrMethod.Write(w)
			log.Info("http method not supported", log15.Ctx{"requestMethod": r.Method})
		}
	})
}

// handler to display a specific existing principal
func (a *AdminAPI) PrincipalGetRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{"method": "PrincipalGetRequest"})

		// get principal by name
		vars := mux.Vars(r)
		principalName := vars["name"]

		log = log.New(log15.Ctx{"principalName": principalName})

		db := a.ctx.PrincipalDB(service.ReadOnly)
		pr, err := principal.PrincipalByNameDB(db, principalName)
		if err == principal.ErrPrincipalNotFound {
			ErrNotFound.Write(w)
			log.Info("principal not found")
			return
		}
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("DB get by name failed", log15.Ctx{"err": err})
			return
		}
		md, err := metadata.MetadataByPrimaryDB(db, principal.MetadataModel, pr.ID)
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("get metadata failed", log15.Ctx{"err": err})
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
	log := a.log.New(log15.Ctx{"method": "putNewPrincipal"})

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

	log = log.New(log15.Ctx{"principalName": pr.Name})

	auth, err := getAuthContainer(r)
	if err != nil {
		log.Crit("context auth error", log15.Ctx{"err": err})
		ErrSystem.Write(w)
		return
	}
	pr.CreatedBy = auth[AuthUserIDKey].(string)
	// set created time
	pr.Created = time.Now()

	// insert pr if not exists
	// start dbtx to save metadata and pr together
	tx, err := a.ctx.PrincipalDB().Begin()
	if err != nil {
		log.Crit("Tx begin failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	// check if principal exists
	_, err = principal.PrincipalByNameTx(tx, pr.Name)
	if err != nil && err != principal.ErrPrincipalNotFound {
		// other db error
		log.Error("DB get by name failed", log15.Ctx{"err": err})
		ErrSystem.Write(w)
		return
	} else if err == nil {
		// already exists
		log.Warn("principal already exists")
		ErrConflict.Write(w)
		return
	}

	// insert pr
	err = principal.InsertPrincipalTx(tx, &pr)
	if err != nil {
		tx.Rollback()
		ErrDatabase.Write(w)
		log.Error("TX insert failed.", log15.Ctx{"err": err})
		return
	}

	// insert metadata
	md := metadata.MetadataFromValues(pr.Metadata, pr.CreatedBy)
	err = metadata.InsertMetadataTx(tx, principal.MetadataModel, pr.ID, md)
	if err != nil {
		tx.Rollback()
		log.Error("TX metadata insert failed.", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	//commit tx
	err = tx.Commit()
	if err != nil {
		ErrDatabase.Write(w)
		log.Crit("TX commit failed.", log15.Ctx{"err": err})
		return
	}

	// get data explicit from DB
	db := a.ctx.PrincipalDB(service.ReadOnly)
	pr, err = principal.PrincipalByNameDB(db, pr.Name)
	if err != nil {
		log.Error("DB get by name failed.", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	md, err = metadata.MetadataByPrimaryDB(db, principal.MetadataModel, pr.ID)
	if len(md) > 0 {
		pr.Metadata = md.Values()
	}
	if err != nil {
		log.Error("get metadata failed.", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	resp := PrincipalAdminAPIResponse{}
	resp.HttpStatus = http.StatusOK
	resp.Status = StatusSuccess
	resp.Response = pr
	resp.Info = "principal " + pr.Name + " created"
	err = resp.Write(w)
	if err != nil {
		log.Error("write error", log15.Ctx{"err": err})
		ErrSystem.Write(w)
		return
	}
}

// post method to add and change the metadata
func (a *AdminAPI) postChangePrincipal(w http.ResponseWriter, r *http.Request) {
	log := a.log.New(log15.Ctx{"method": "postChangePrincipal"})

	// get Metadata from post variables
	jd := json.NewDecoder(r.Body)
	pr := principal.Principal{}
	err := jd.Decode(&pr)
	r.Body.Close()
	if err != nil {
		log.Error("json decode failed", log15.Ctx{"err": err})
		ErrReadJson.Write(w)
		return
	}

	auth, err := getAuthContainer(r)
	if err != nil {
		log.Crit("error getting auth container", log15.Ctx{"err": err})
		ErrSystem.Write(w)
		return
	}
	pr.CreatedBy = auth[AuthUserIDKey].(string)

	log = log.New(log15.Ctx{"principalName": pr.Name})

	// open transaction to add the posted metadata
	tx, err := a.ctx.PrincipalDB().Begin()
	if err != nil {
		log.Crit("Tx begin failed.", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	// does principal exist
	pr.ID, err = principal.PrincipalIDByNameTx(tx, pr.Name)
	if err != nil {
		txErr := tx.Rollback()
		if txErr != nil {
			log.Crit("error on rollback", log15.Ctx{"err": err})
		}
		if err == principal.ErrPrincipalNotFound {
			log.Warn("principal not found")
			ErrNotFound.Write(w)
			return
		}
		log.Error("get principal from DB failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	md := metadata.MetadataFromValues(pr.Metadata, pr.CreatedBy)

	err = metadata.InsertMetadataTx(tx, principal.MetadataModel, pr.ID, md)
	if err != nil {
		txErr := tx.Rollback()
		if txErr != nil {
			log.Crit("error on rollback", log15.Ctx{"err": err})
		}
		log.Error("insert metadata failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	err = tx.Commit()
	if err != nil {
		log.Crit("error on commit", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	// get stored data from db
	db := a.ctx.PrincipalDB(service.ReadOnly)
	pr, err = principal.PrincipalByNameDB(db, pr.Name)
	if err != nil {
		if err == principal.ErrPrincipalNotFound {
			log.Error("principal not found")
			ErrNotFound.Write(w)
			return
		}
		log.Error("DB get by name failed.", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	md, err = metadata.MetadataByPrimaryDB(db, principal.MetadataModel, pr.ID)
	if err != nil {
		log.Error("get metadata failed.", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	if len(md) > 0 {
		pr.Metadata = md.Values()
	}

	// create response
	resp := PrincipalAdminAPIResponse{}
	resp.Info = "principal " + pr.Name + " changed"
	resp.Status = StatusSuccess
	resp.Response = pr
	err = resp.Write(w)
	if err != nil {
		ErrSystem.Write(w)
		log.Error("write error", log15.Ctx{"err": err})
		return
	}
}
