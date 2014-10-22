package v1

import (
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"path"
)

const (
	APIParamPrincipalName = "principalName"
)

func (a *AdminAPI) PrincipalRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		log := a.log.New(log15.Ctx{"method": "Principal Request"})
		log.Info("Method:" + r.Method)

		if r.Method == "GET" {
			// get principal by name
			urlpath, principalName := path.Split(path.Clean(r.URL.Path))

			if len(principalName) < 1 {
				w.WriteHeader(http.StatusBadRequest)
				log.Info("principalName missing")
				return
			}
			log.Info("principalName: " + principalName)
			log.Info("url path: " + urlpath)

			pr, err := principal.PrincipalByNameDB(a.ctx.PrincipalDB(service.ReadOnly), principalName)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				log.Error("DB get by name failed.", log15.Ctx{"err": err})
				return
			}
			je := json.NewEncoder(w)
			err = je.Encode(&pr)

		} else if r.Method == "PUT" {
			// create new principal
			jd := json.NewDecoder(r.Body)
			pr := principal.Principal{}
			err := jd.Decode(&pr)
			r.Body.Close()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				log.Error("json decode failed", log15.Ctx{"err": err})

				return
			}

			db := a.ctx.PrincipalDB()
			_, err = principal.PrincipalByNameDB(db, pr.Name)
			if err == principal.ErrPrincipalNotFound {
				// insert if not exists
				err = principal.InsertPrincipalDB(db, &pr)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					log.Error("DB insert failed.", log15.Ctx{"err": err})
					return
				}
			} else if err != nil {
				// other db error
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("DB get by name failed.", log15.Ctx{"err": err})
				return
			} else {
				// already exists
				w.WriteHeader(http.StatusConflict)
				log.Warn("principal already exist: " + pr.Name)
				return
			}

			pr, err = principal.PrincipalByNameDB(db, pr.Name)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("DB get by name failed.", log15.Ctx{"err": err})
				return
			}

			je := json.NewEncoder(w)
			err = je.Encode(&pr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("json encode failed.", log15.Ctx{"err": err})
				return
			}

		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})
}
