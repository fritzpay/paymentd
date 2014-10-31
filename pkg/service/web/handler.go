package web

import (
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
)

type Handler struct {
	ctx *service.Context
	log log15.Logger

	router *mux.Router
}

func NewHandler(ctx *service.Context) (*Handler, error) {
	h := &Handler{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/web",
		}),

		router: mux.NewRouter(),
	}
	return h, nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			h.log.Crit("panic on serving HTTP", log15.Ctx{"panic": err})
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	service.SetRequestContext(r, h.ctx)
	defer service.ClearRequestContext(r)
	h.router.ServeHTTP(w, r)
}
