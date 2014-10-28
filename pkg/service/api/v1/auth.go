package v1

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"hash"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/fritzpay/paymentd/pkg/paymentd/config"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
)

const badAuthWaitTime = 2 * time.Second

func (a *AdminAPI) authorizationHash() func() hash.Hash {
	return sha256.New
}

func (a *AdminAPI) authenticateSystemPassword(pw string, w http.ResponseWriter) {
	log := a.log.New(log15.Ctx{"method": "authenticateSystemPassword"})
	pwEntry, err := config.EntryByNameDB(a.ctx.PaymentDB(), config.ConfigNameSystemPassword)
	if err != nil {
		if err == config.ErrEntryNotFound {
			log.Error("no password entry")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		log.Error("error retrieving password entry", log15.Ctx{"err": err})
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
		return
	}
	a.respondWithAuthorization(w)
}

// GetCredentialsResponse is the response for all GET /user/credentials requests
// ready to be JSON-encoded
type GetCredentialsResponse struct {
	Authorization string
}

func (a *AdminAPI) respondWithAuthorization(w http.ResponseWriter) {
	log := a.log.New(log15.Ctx{"method": "respondWithAuthorization"})

	auth := service.NewAuthorization(a.authorizationHash())
	auth.Payload[AuthUserIDKey] = systemUserID
	auth.Expires(time.Now().Add(AuthLifetime))
	key, err := a.ctx.Keychain().BinKey()
	if err != nil {
		log.Error("error retrieving key from keychain", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = auth.Encode(key)
	if err != nil {
		log.Error("error encoding authorization", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	resp := GetCredentialsResponse{}
	resp.Authorization, err = auth.Serialized()
	if err != nil {
		log.Error("error serializing authorization", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Error("error encoding JSON respone", log15.Ctx{"err": err})
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if a.ctx.Config().API.Cookie.AllowCookieAuth {
		c := &http.Cookie{
			Name:     AuthCookieName,
			Value:    resp.Authorization,
			Path:     ServicePath,
			Expires:  auth.Expiry(),
			HttpOnly: a.ctx.Config().API.Cookie.HTTPOnly,
			Secure:   a.ctx.Config().API.Cookie.Secure,
		}
		http.SetCookie(w, c)
	}
	_, err = w.Write(jsonResp)
	if err != nil {
		log.Error("error writing HTTP response", log15.Ctx{"err": err})
	}
}

// AuthorizationHandler implements /authorization requests
func (a *AdminAPI) AuthorizationHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET":
			a.AuthRequiredHandler(a.refreshAuthorizationHandler()).ServeHTTP(w, r)

		case "POST":
			a.AuthRequiredHandler(a.updateSystemUserPasswordHandler()).ServeHTTP(w, r)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

// AuthorizeHandler handles new authorizations
func (a *AdminAPI) AuthorizeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "GET":
			vars := mux.Vars(r)
			authMethod := vars["method"]
			switch authMethod {
			case "basic":
				a.authenticateBasicAuth(w, r)
				return
			default:
				w.WriteHeader(http.StatusNotFound)
				return
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
	})
}

func getAuthorizationMethod(p string) string {
	_, method := path.Split(path.Clean(p))
	return method
}

func getBasicAuthPassword(authHeader string) (string, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 {
		return "", errors.New("authorization expect two parts")
	}
	if parts[0] != "Basic" {
		return "", errors.New("not basic auth")
	}
	auth, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	parts = strings.Split(string(auth), ":")
	if len(parts) != 2 {
		return "", errors.New("password expect two parts")
	}
	return parts[1], nil
}

func (a *AdminAPI) authenticateBasicAuth(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") == "" {
		requestBasicAuth(w)
		return
	}
	if pw, err := getBasicAuthPassword(r.Header.Get("Authorization")); err != nil {
		a.log.Warn("error on basic auth", log15.Ctx{"err": err})
		requestBasicAuth(w)
		return
	} else {
		a.authenticateSystemPassword(pw, w)
	}
}

func requestBasicAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization\"")
	w.WriteHeader(http.StatusUnauthorized)
}

func (a *AdminAPI) updateSystemUserPasswordHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := a.log.New(log15.Ctx{"method": "updateSystemUserPasswordHandler"})
		w.Header().Set("Content-Type", "text/plain")
		if !strings.Contains(r.Header.Get("Content-Type"), "text/plain") {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}
		pw, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Error("error reading request body", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = config.Set(a.ctx.PaymentDB(), config.SetPassword(pw))
		if err != nil {
			log.Error("error setting system password", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func (a *AdminAPI) refreshAuthorizationHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.respondWithAuthorization(w)
	})
}

// AuthRequiredHandler wraps the given handler with an authorization method using the
// Authorization Header and the authorization container
//
// A failed authorization will lead to a http.StatusUnauthorized header
func (a *AdminAPI) AuthRequiredHandler(parent http.Handler) http.Handler {
	return a.AuthHandler(parent, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
}

// AuthHandler wraps the given handler with an authorization method using the
// Authorization Header and the authorization container
//
// When the request can be authorized, the success handler will be called, otherwise
// the failed handler will be called
func (a *AdminAPI) AuthHandler(success, failed http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := a.log.New(log15.Ctx{"method": "AuthHandler"})

		authStr := r.Header.Get("Authorization")
		if authStr == "" {
			if !a.ctx.Config().API.Cookie.AllowCookieAuth {
				if Debug {
					log.Debug("missing authorization header")
				}
				failed.ServeHTTP(w, r)
				return
			}
			c, err := r.Cookie(AuthCookieName)
			if err != nil {
				if err != http.ErrNoCookie {
					log.Warn("error retrieving auth cookie", log15.Ctx{"err": err})
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// ErrNoCookie
				failed.ServeHTTP(w, r)
				return
			}
			authStr = c.Value
		}
		auth := service.NewAuthorization(a.authorizationHash())
		_, err := auth.ReadFrom(strings.NewReader(authStr))
		if err != nil {
			if Debug {
				log.Debug("error reading authorization", log15.Ctx{"err": err})
			}
			failed.ServeHTTP(w, r)
			return
		}
		if auth.Expiry().Before(time.Now()) {
			if Debug {
				log.Debug("authorization expired", log15.Ctx{"expiry": auth.Expiry()})
			}
			failed.ServeHTTP(w, r)
			return
		}
		key, err := a.ctx.Keychain().MatchKey(auth)
		if err != nil {
			if Debug {
				log.Debug("error retrieving matching key from keychain", log15.Ctx{
					"err":            err,
					"keysInKeychain": a.ctx.Keychain().KeyCount(),
				})
			}
			failed.ServeHTTP(w, r)
			return
		}
		err = auth.Decode(key)
		if err != nil {
			if Debug {
				log.Debug("error decoding authorization", log15.Ctx{"err": err})
			}
			failed.ServeHTTP(w, r)
			return
		}
		// store auth container in request context
		service.SetRequestContextVar(r, service.ContextVarAuthKey, auth.Payload)

		success.ServeHTTP(w, r)
	})
}

func getAuthContainer(r *http.Request) (map[string]interface{}, error) {
	ctx := service.RequestContext(r)
	if ctx == nil {
		return nil, errors.New("request context not present")
	}
	auth, ok := ctx.Value(service.ContextVarAuthKey).(map[string]interface{})
	if !ok {
		return nil, errors.New("auth container type error")
	}
	return auth, nil
}
