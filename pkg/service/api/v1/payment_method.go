package v1

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fritzpay/paymentd/pkg/metadata"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/paymentd/provider"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

// PaymentMethodRequest is the request JSON struct for POST - PUT
// project/(id)/method/
type PaymentMethodRequest struct {
	MethodKey string
	Provider  string
	Status    string
	CreatedBy string
	Metadata  map[string]string
}

func (a *AdminAPI) PaymentMethodGetRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log := a.log.New(log15.Ctx{"method": "Project payment methods GET"})

		// parameter
		vars := mux.Vars(r)
		projectIDParam := vars["projectid"]
		methodKey := vars["methodkey"]
		providerParam := vars["provider"]

		projectID, err := strconv.ParseInt(projectIDParam, 10, 64)
		if err != nil {
			ErrReadParam.Write(w)
			log.Error("param conversion error", log15.Ctx{"err": err})
			return
		}
		if methodKey == "" {
			ErrInval.Write(w)
			log.Error("param conversion error", log15.Ctx{"methodkey": methodKey})
			return
		}

		// get payment method
		db := a.ctx.PaymentDB(service.ReadOnly)
		pm, err := payment_method.PaymentMethodByProjectIDProviderNameMethodKeyDB(db, projectID, providerParam, methodKey)
		if err == payment_method.ErrPaymentMethodNotFound {
			ErrNotFound.Write(w)
			log.Error("error retrieving payment method", log15.Ctx{"err": err})
			return
		}
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("database error", log15.Ctx{"err": err})
			return
		}

		// return methods
		resp := ProjectAdminAPIResponse{}
		resp.HttpStatus = http.StatusOK
		resp.Info = "paymentmethod found"
		resp.Response = pm
		resp.Write(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("write error", log15.Ctx{"err": err})
			return
		}
	})
}

// handler to create or change a principal
//
// PUT creates new principal
// POST can be used to change the principals metadata
func (a *AdminAPI) PaymentMethodRequest() http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{"method": "PaymentMethodRequest"})

		log.Info("project method", log15.Ctx{"method": r.Method})

		switch r.Method {
		case "PUT":
			a.putNewPaymentMethod(w, r)
		case "POST":
			a.postChangePaymentMethod(w, r)
		default:
			ErrMethod.Write(w)
			log.Info("http method not supported", log15.Ctx{"requestMethod": r.Method})
		}
	})
	return a.ctx.RateLimitHandler(h)
}

func (a *AdminAPI) putNewPaymentMethod(w http.ResponseWriter, r *http.Request) {
	log := a.log.New(log15.Ctx{"method": "PaymentMethod PUT Request"})
	// get parameters
	// projectid and methodname
	vars := mux.Vars(r)
	projectIDParam := vars["projectid"]

	log.Info("put project", log15.Ctx{"projectID": projectIDParam})

	projectID, err := strconv.ParseInt(projectIDParam, 10, 64)
	if err != nil {
		ErrReadParam.Write(w)
		log.Info("malformed param", log15.Ctx{"projectIdParam": projectIDParam})
		return
	}

	db := a.ctx.PrincipalDB()
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("error on begin", log15.Ctx{"err": err})
	}

	proj, err := project.ProjectByIDDB(db, projectID)
	if err != nil && err != project.ErrProjectNotFound {
		ErrDatabase.Write(w)
		log.Error("database request failed", log15.Ctx{"err": err})
		return
	}
	if err == project.ErrProjectNotFound {
		ErrNotFound.Write(w)
		log.Warn("project does not exist", log15.Ctx{"err": err})
		return
	}

	// parse request
	jd := json.NewDecoder(r.Body)
	pmr := PaymentMethodRequest{}
	err = jd.Decode(&pmr)
	if err != nil {
		ErrReadJson.Write(w)
		log.Error("json decoding failed", log15.Ctx{"err": err})
		return
	}
	r.Body.Close()

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
	// get Provider
	tx, err = a.ctx.PaymentDB().Begin()
	prov, err := provider.ProviderByNameTx(tx, pmr.Provider)
	if err != nil {
		commit = true
		ErrDatabase.Write(w)
		log.Error("database request failed", log15.Ctx{"err": err})
		return
	}

	// set paymentMethod values
	// get user id
	auth := service.RequestContextAuth(r)
	var pm payment_method.Method
	// parse status value
	pm.Status, err = payment_method.ParseMethodStatus(pmr.Status)
	if err != nil {
		ErrInval.Write(w)
		return
	}
	pm.Created = time.Now().UTC().Round(time.Second)
	pm.CreatedBy = auth[AuthUserIDKey].(string)
	pm.StatusCreatedBy = pm.CreatedBy
	pm.ProjectID = proj.ID
	pm.Provider = prov
	pm.MethodKey = pmr.MethodKey
	pm.Metadata = pmr.Metadata

	// check if payment_method already exists
	_, err = payment_method.PaymentMethodByProjectIDProviderNameMethodKeyTx(tx, pm.ProjectID, pm.Provider.Name, pm.MethodKey)
	if err != nil && err != payment_method.ErrPaymentMethodNotFound {
		ErrDatabase.Write(w)
		log.Error("database error", log15.Ctx{"err": err})
		return
	}
	if err == nil {
		ErrConflict.Write(w)
		log.Error("conflict", log15.Ctx{"err": err})
		return
	}
	// insert method
	err = payment_method.InsertPaymentMethodTx(tx, &pm)
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("database error", log15.Ctx{"err": err})
		return
	}

	// insert status
	err = payment_method.InsertPaymentMethodStatusTx(tx, &pm)
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("database error", log15.Ctx{"err": err})
		return
	}

	// insert method metadata
	md := metadata.MetadataFromValues(pm.Metadata, pm.CreatedBy)
	err = metadata.InsertMetadataTx(tx, payment_method.MetadataModel, pm.ID, md)
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("database error", log15.Ctx{"err": err})
		return
	}
	if err != nil && err != payment_method.ErrPaymentMethodNotFound {
		ErrDatabase.Write(w)
		log.Error("database error", log15.Ctx{"err": err})
		return
	}

	// get payment_method from db with all set values like status created
	pmdb, err := payment_method.PaymentMethodByProjectIDProviderNameMethodKeyTx(tx, pm.ProjectID, pm.Provider.Name, pm.MethodKey)
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("database error", log15.Ctx{"err": err})
		return
	}

	commit = true
	err = tx.Commit()
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("database error", log15.Ctx{"err": err})
		return
	}

	resp := ProjectAdminAPIResponse{}
	resp.Status = StatusSuccess
	resp.Info = "created with methodkey " + pmr.MethodKey
	resp.Response = pmdb
	err = resp.Write(w)
	if err != nil {
		log.Error("error writing response", log15.Ctx{"err": err})
	}
}

func (a *AdminAPI) postChangePaymentMethod(w http.ResponseWriter, r *http.Request) {
	log := a.log.New(log15.Ctx{"method": "PaymentMethod POST Request"})
	// get parameters
	// projectid and methodname
	vars := mux.Vars(r)
	projectIDParam := vars["projectid"]
	methodKey := vars["methodkey"]

	projectID, err := strconv.ParseInt(projectIDParam, 10, 64)
	if err != nil {
		ErrReadParam.Write(w)
		log.Info("malformed param", log15.Ctx{"projectIdParam": projectIDParam})
		return
	}
	if methodKey == "" {
		ErrReadParam.Write(w)
		log.Info("malformed param", log15.Ctx{"methodKey missing": methodKey})
		return
	}

	// parse request
	jd := json.NewDecoder(r.Body)
	pmr := PaymentMethodRequest{}
	err = jd.Decode(&pmr)
	if err != nil {
		ErrReadJson.Write(w)
		log.Error("json decoding failed", log15.Ctx{"err": err})
		return
	}
	r.Body.Close()

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

	tx, err = a.ctx.PaymentDB().Begin()
	// check if payment_method exists
	pm, err := payment_method.PaymentMethodByProjectIDProviderNameMethodKeyTx(tx, projectID, pmr.Provider, methodKey)
	if err == payment_method.ErrPaymentMethodNotFound {
		ErrNotFound.Write(w)
		log.Error("payment method not found", log15.Ctx{"err": err})
		return
	}
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("database error", log15.Ctx{"err": err})
		return
	}

	// user data
	auth := service.RequestContextAuth(r)
	// insert new status if set
	if pmr.Status != pm.Status.String() {
		// set paymentMethod values
		// get user id
		// parse status value
		pm.Status, err = payment_method.ParseMethodStatus(pmr.Status)
		if err != nil {
			ErrInval.Write(w)
			return
		}
		pm.Status.Scan(pmr.Status)
		pm.StatusCreatedBy = auth[AuthUserIDKey].(string)
		payment_method.InsertPaymentMethodStatusTx(tx, pm)
		// reload payment_method
		pm, err = payment_method.PaymentMethodByProjectIDProviderNameMethodKeyTx(tx, pm.ProjectID, pm.Provider.Name, pm.MethodKey)
		if err == payment_method.ErrPaymentMethodNotFound {
			ErrNotFound.Write(w)
			log.Error("payment method not found", log15.Ctx{"err": err})
			return
		}

		if err != nil {
			ErrDatabase.Write(w)
			log.Error("database error", log15.Ctx{"err": err})
			return
		}
	}
	// insert metadata if set
	if pmr.Metadata != nil {
		md := metadata.MetadataFromValues(pmr.Metadata, auth[AuthUserIDKey].(string))
		err = metadata.InsertMetadataTx(tx, payment_method.MetadataModel, pm.ID, md)
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("database error", log15.Ctx{"err": err})
			return
		}
		if err != nil && err != payment_method.ErrPaymentMethodNotFound {
			ErrDatabase.Write(w)
			log.Error("database error", log15.Ctx{"err": err})
			return
		}

		// reload payment method matadata
		pmmd, err := payment_method.PaymentMethodMetadataTx(tx, pm)
		if err != nil {
			ErrDatabase.Write(w)
			log.Error("database error", log15.Ctx{"err": err})
			return
		}
		pm.Metadata = pmmd
	}

	resp := ProjectAdminAPIResponse{}
	resp.Status = StatusSuccess
	resp.Info = "changed " + methodKey
	resp.Response = pm
	resp.Write(w)

	err = tx.Commit()
	if err != nil {
		ErrDatabase.Write(w)
		log.Error("database error", log15.Ctx{"err": err})
		return
	}
	commit = true
}
