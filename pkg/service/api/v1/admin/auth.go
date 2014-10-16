package admin

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"hash"
	"net/http"
	"path"
	"strings"
	"time"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/fritzpay/paymentd/pkg/paymentd/config"
	"github.com/fritzpay/paymentd/pkg/service"
	"gopkg.in/inconshreveable/log15.v2"
)

func (a *API) authorizationHash() func() hash.Hash {
	return sha256.New
}

func (a *API) authenticateSystemPassword(pw string, w http.ResponseWriter) {
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

func (a *API) respondWithAuthorization(w http.ResponseWriter) {
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
			Expires:  auth.Expiry(),
			HttpOnly: a.ctx.Config().API.Cookie.HttpOnly,
			Secure:   a.ctx.Config().API.Cookie.Secure,
		}
		if servicePath, ok := a.ctx.Value("ServicePath").(string); ok {
			c.Path = servicePath
		}
		http.SetCookie(w, c)
	}
	_, err = w.Write(jsonResp)
	if err != nil {
		log.Error("error writing HTTP response", log15.Ctx{"err": err})
	}
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
		a.authenticateBasicAuth(w, r)
		return
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
}

func (a *API) authenticateBasicAuth(w http.ResponseWriter, r *http.Request) {
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
}

func requestBasicAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", "Basic realm=\"Authorization\"")
	w.WriteHeader(http.StatusUnauthorized)
}

func getCredentialsMethod(p string) string {
	_, method := path.Split(path.Clean(p))
	return method
}

// AuthHandler wraps the given handler with an authorization method using the
// Authorization Header and the authorization container
func (a *API) AuthHandler(parent http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := a.log.New(log15.Ctx{"method": "AuthHandler"})

		authStr := r.Header.Get("Authorization")
		if authStr == "" {
			if !a.ctx.Config().API.Cookie.AllowCookieAuth {
				log.Debug("missing authorization header")
				w.WriteHeader(http.StatusUnauthorized)
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
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			authStr = c.Value
		}
		auth := service.NewAuthorization(a.authorizationHash())
		_, err := auth.ReadFrom(strings.NewReader(authStr))
		if err != nil {
			log.Debug("error reading authorization", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if auth.Expiry().Before(time.Now()) {
			log.Debug("authorization expired", log15.Ctx{"expiry": auth.Expiry()})
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		key, err := a.ctx.Keychain().MatchKey(auth)
		if err != nil {
			log.Debug("error retrieving matching key from keychain", log15.Ctx{
				"err":            err,
				"keysInKeychain": a.ctx.Keychain().KeyCount(),
			})
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		err = auth.Decode(key)
		if err != nil {
			log.Debug("error decoding authorization", log15.Ctx{"err": err})
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if userID, ok := auth.Payload[AuthUserIDKey]; !ok {
			log.Debug("no userID present")
			w.WriteHeader(http.StatusUnauthorized)
			return
		} else {
			service.SetRequestContextVar(r, AuthUserIDKey, userID)
		}

		parent.ServeHTTP(w, r)
	})
}
