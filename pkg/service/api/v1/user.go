package v1

import (
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

type UserAdminAPIResponse struct {
	AdminAPIResponse
}

// GetUserID returns a utility handler. This endpoint displays the user ID, which is stored
// in the authorization container
func (a *AdminAPI) GetUserID() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method != "GET" {
			ErrMethod.Write(w)
			return
		}
		log := a.log.New(log15.Ctx{"method": "GetUserID"})

		auth, err := getAuthContainer(r)
		if err != nil {
			log.Crit("auth container error", log15.Ctx{"err": err})
			ErrSystem.Write(w)
			return
		}

		resp := UserAdminAPIResponse{}
		resp.Info = "user id"
		resp.Status = StatusSuccess
		resp.Response = auth[AuthUserIDKey].(string)
		resp.Write(w)
	})
}
