package v1

import (
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/metadata"
	"github.com/fritzpay/paymentd/pkg/paymentd/payment_method"
	"github.com/fritzpay/paymentd/pkg/paymentd/project"
	"github.com/fritzpay/paymentd/pkg/paymentd/provider"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"strconv"
	"time"
)

// PaymentMethodRequest is the request JSON struct for POST - PUT
// project/(id)/method/
type PaymentMethodRequest struct {
	MethodKey  string
	ProviderID int64 `json:",string"`
	Status     string
	CreatedBy  string
	Metadata   map[string]string
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
			log.Error("database error", log15.Ctx{"err": err})
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
		w.Header().Set("Content-Type", "application/json")
		log := a.log.New(log15.Ctx{"method": "Project payment method request"})

		// PUT create new entry
		if r.Method == "PUT" {
			// get parameters
			// projectid and methodname
			// check parameters exits in db
			vars := mux.Vars(r)
			projectIdParam := vars["projectid"]

			projectId, err := strconv.ParseInt(projectIdParam, 10, 64)
			if err != nil {
				ErrReadParam.Write(w)
				log.Info("malformed param", log15.Ctx{"projectIdParam": projectIdParam})
				return
			}
			db := a.ctx.PrincipalDB()
			proj, err := project.ProjectByIdDB(db, projectId)
			if err != nil {
				ErrDatabase.Write(w)
				log.Error("database request failed", log15.Ctx{"err": err})
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

			// get Provider
			paydb := a.ctx.PaymentDB()
			prov, err := provider.ProviderByIDDB(paydb, pmr.ProviderID)
			if err != nil {
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
			pm.StatusChanged = time.Now()
			pm.StatusCreatedBy = auth[AuthUserIDKey].(string)
			if err != nil {
				ErrInval.Write(w)
				return
			}
			pm.ProjectID = proj.ID
			pm.Provider = prov
			pm.MethodKey = pmr.MethodKey
			pm.Created = time.Now()
			pm.CreatedBy = auth[AuthUserIDKey].(string)
			pm.Metadata = pmr.Metadata

			// check if payment_method already exists
			pmdb, err := payment_method.PaymentMethodByProjectIDProviderIDMethodKey(paydb, pm.ProjectID, pm.Provider.ID, pm.MethodKey)
			if err == payment_method.ErrPaymentMethodNotFound {

				// save to db
				tx, err := paydb.Begin()
				if err != nil {
					ErrDatabase.Write(w)
					log.Error("database error", log15.Ctx{"err": err})
				}
				paymentMethodId, err := payment_method.InsertPaymentMethodTx(tx, pm)
				if err != nil {
					tx.Rollback()
					ErrDatabase.Write(w)
					log.Error("database error", log15.Ctx{"err": err})
					return
				}
				pm.ID = paymentMethodId

				// save method metadata
				md := metadata.MetadataFromValues(pm.Metadata, pm.CreatedBy)
				err = metadata.InsertMetadataTx(tx, payment_method.MetadataModel, pm.ID, md)
				if err != nil {
					tx.Rollback()
					ErrDatabase.Write(w)
					log.Error("database error", log15.Ctx{"err": err})
					return
				}
				err = tx.Commit()
				if err != nil {
					ErrDatabase.Write(w)
					log.Error("database error", log15.Ctx{"err": err})
					return
				}

				resp := ProjectAdminAPIResponse{}
				resp.Status = StatusSuccess
				resp.Info = "created with methodkey " + pmr.MethodKey
				resp.Response = pm
				resp.Write(w)
			} else if err != nil {
				ErrConflict.Write(w)
				log.Error("conflict", log15.Ctx{"err": err})
				return
			} else {
				ErrConflict.Write(w)
				log.Info("conflict", log15.Ctx{"err": err})
				log.Info("conflict", log15.Ctx{"MethodKey": pmdb.ProjectID})
				log.Info("conflict", log15.Ctx{"MethodKey": pmdb.Provider.ID})
				log.Info("conflict", log15.Ctx{"MethodKey": pmdb.MethodKey})
				return
			}

		} else if r.Method == "POST" {

			// POST change
			// set status
			// set metadata

		} else {
			ErrMethod.Write(w)
			log.Info("unsupported method", log15.Ctx{"requestMethod": r.Method})
			return
		}

	})
}
