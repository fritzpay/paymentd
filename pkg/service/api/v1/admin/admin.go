package admin

import (
	"github.com/gorilla/mux"
	"net/http"
)

type API struct {
	router *mux.Router
}

func NewAPI(r *mux.Router) *API {
	a := &API{
		router: r,
	}

	a.router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test here"))
	})
	return a
}
