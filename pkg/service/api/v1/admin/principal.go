package admin

import (
	"encoding/json"
	"github.com/fritzpay/paymentd/pkg/paymentd/principal"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"path"
	"strings"
)

const (
	APIParamPrincipalName = "principalName"
)

func (a *API) PrincipalRequest() http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		log := a.log.New(log15.Ctx{"method": "Principal Request"})
		log.Info("Method:" + r.Method)

		if r.Method == "GET" {
			// get principal by name
			principalName := strings.TrimLeft(r.RequestURI, path.Dir(r.RequestURI))
			if len(principalName) < 1 {
				w.WriteHeader(http.StatusInternalServerError)
				log.Info("principalName missing")
				return
			}
			log.Info("Param principalName " + principalName)

			pr, err := principal.PrincipalByNameDB(a.ctx.PrincipalDB(), principalName)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
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
			err = principal.InsertPrincipalDB(db, &pr)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				log.Error("DB insert failed.", log15.Ctx{"err": err})
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
