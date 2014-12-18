package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/fritzpay/paymentd/pkg/metadata"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

type PrincipalAdminAPIResponse struct {
	AdminAPIResponse
}

// handler to create or change a principal
//
// PUT creates new principal
// POST can be used to change the principals metadata
func (a *AdminAPI) PrincipalRequest() http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{"method": "PrincipalRequest"})

		switch r.Method {
		case "PUT":
			a.putNewPrincipal(w, r)
		default:
			ErrMethod.Write(w)
			log.Info("http method not supported", log15.Ctx{"requestMethod": r.Method})
		}
	})
	return a.ctx.RateLimitHandler(h)
}

func (a *AdminAPI) PrincipalNameRequest() http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			a.ctx.RateLimitHandler(http.HandlerFunc(a.postChangePrincipal)).ServeHTTP(w, r)
		case "GET":
			a.getPrincipal(w, r)
		default:
			ErrMethod.Write(w)
		}
	})
	return h
}

// handler to display a specific existing principal
func (a *AdminAPI) getPrincipal(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	log := a.log.New(log15.Ctx{"method": "getPrincipal"})

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
	if err := pr.ValidStatus(); err != nil {
		ErrInval.Write(w)
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
	pr.Created = time.Now().UTC().Round(time.Second)

	// insert pr if not exists
	// start dbtx to save metadata and pr together
	tx, err := a.ctx.PrincipalDB().Begin()
	if err != nil {
		log.Crit("Tx begin failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	// insert principal
	err = principal.InsertPrincipalTx(tx, &pr)
	if err != nil {
		tx.Rollback()
		ErrDatabase.Write(w)
		log.Error("TX insert failed.", log15.Ctx{"err": err})
		return
	}
	// insert principal status
	err = principal.InsertPrincipalStatusTx(tx, pr, pr.CreatedBy)
	if err != nil {
		tx.Rollback()
		ErrDatabase.Write(w)
		log.Error("error on insert principal status", log15.Ctx{"err": err})
		return
	}
	// insert Metadata
	err = insertPrincipalMetadata(tx, &pr)
	if err != nil {
		tx.Rollback()
		ErrDatabase.Write(w)
		log.Error("metadata insert failed.", log15.Ctx{"err": err})
		return
	}

	//commit tx
	err = tx.Commit()
	if err != nil {
		ErrDatabase.Write(w)
		log.Crit("TX commit failed.", log15.Ctx{"err": err})
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

	vars := mux.Vars(r)
	principalName := vars["name"]

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
	pr.Created = time.Now().UTC().Round(time.Second)

	log = log.New(log15.Ctx{"principalID": pr.ID})

	// open transaction to add the posted metadata
	tx, err := a.ctx.PrincipalDB().Begin()
	if err != nil {
		log.Crit("Tx begin failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}

	// does principal exist
	prByName, err := principal.PrincipalByNameTx(tx, principalName)
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
	pr.ID = prByName.ID

	// insert Metadata
	md := metadata.MetadataFromValues(pr.Metadata, pr.CreatedBy)
	err = insertPrincipalMetadata(tx, &pr)
	if err != nil {
		tx.Rollback()
		ErrDatabase.Write(w)
		log.Error("metadata insert failed", log15.Ctx{"err": err})
		return
	}

	// get stored and added metadata from db
	md, err = metadata.MetadataByPrimaryTx(tx, principal.MetadataModel, pr.ID)
	if err != nil {
		tx.Rollback()
		log.Error("get metadata failed", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
	}
	if len(md) > 0 {
		pr.Metadata = md.Values()
	}
	err = tx.Commit()
	if err != nil {
		log.Crit("error on commit", log15.Ctx{"err": err})
		ErrDatabase.Write(w)
		return
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

// adds metadata into the database
func insertPrincipalMetadata(tx *sql.Tx, pr *principal.Principal) error {

	// check if principal exists
	_, err := principal.PrincipalByNameTx(tx, pr.Name)
	if err != nil && err != principal.ErrPrincipalNotFound {
		return err
	}

	// insert metadata
	md := metadata.MetadataFromValues(pr.Metadata, pr.CreatedBy)
	err = metadata.InsertMetadataTx(tx, principal.MetadataModel, pr.ID, md)

	return err
}
