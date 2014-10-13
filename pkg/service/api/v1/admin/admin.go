package admin

import (
	"code.google.com/p/go.crypto/bcrypt"
	"encoding/base64"
	"github.com/fritzpay/paymentd/pkg/paymentd/config"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"path"
	"strings"
	"time"
)

const (
	badAuthWaitTime = 2 * time.Second
)

// API represents the admin API in version 1.x
type API struct {
	ctx *service.Context
	log log15.Logger
}

// NewAPI creates a new admin API
func NewAPI(ctx *service.Context) *API {
	a := &API{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1/admin"}),
	}
	return a
}

func (a *API) authenticateSystemPassword(pw string, w http.ResponseWriter) {
	log := a.log.New(log15.Ctx{"method": "authenticateSystemPassword"})
	pwEntry, err := config.EntryByNameDB(a.ctx.PaymentDB(), config.ConfigNameSystemPassword)
	if err != nil {
		log.Error("error retrieving password entry", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if pwEntry == nil {
		log.Error("no password entry")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(pwEntry.Value), []byte(pw))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			time.Sleep(badAuthWaitTime)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		log.Error("error checking password", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Write([]byte("ok"))
}

// GetCredentials implements the GET /user/credentials request
func (a *API) GetCredentials(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	switch getCredentialsMethod(r.URL.Path) {
	case "basic":
		if r.Header.Get("Authorization") == "" {
			requestBasicAuth(w)
			return
		}
		parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
		if parts[0] != "Basic" {
			requestBasicAuth(w)
			return
		}
		auth, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			requestBasicAuth(w)
			return
		}
		parts = strings.Split(string(auth), ":")
		if len(parts) != 2 {
			requestBasicAuth(w)
			return
		}
		a.authenticateSystemPassword(parts[1], w)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func requestBasicAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization\"")
	w.WriteHeader(http.StatusUnauthorized)
}

func getCredentialsMethod(p string) string {
	_, method := path.Split(path.Clean(p))
	return method
}
