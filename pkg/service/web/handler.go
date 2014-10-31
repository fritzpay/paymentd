package web

import (
	"fmt"
	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
	"net/http"
	"os"
)

const (
	PaymentPath = "/payment"
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

	err := h.registerPayment()
	if err != nil {
		h.log.Error("error registering payment", log15.Ctx{"err": err})
		return nil, err
	}

	err = h.registerPublic()
	if err != nil {
		h.log.Error("error registering www public dir", log15.Ctx{"err": err})
		return nil, err
	}

	return h, nil
}

func (h *Handler) requireDir(dir string) error {
	inf, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("error opening dir: %v", err)
	}
	if !inf.IsDir() {
		return fmt.Errorf("dir is not a directory")
	}
	return nil
}

func (h *Handler) registerPayment() error {
	h.log.Info("registering web payment hander...")
	cfg := h.ctx.Config()
	if cfg.Web.TemplateDir == "" {
		return fmt.Errorf("no template dir configured")
	}
	if err := h.requireDir(cfg.Web.TemplateDir); err != nil {
		return fmt.Errorf("error on template dir: %v", err)
	}
	h.router.Handle(PaymentPath, h.PaymentHandler()).Methods("GET")
	return nil
}

func (h *Handler) registerPublic() error {
	h.log.Info("registering www public directory...")
	cfg := h.ctx.Config()
	if cfg.Web.PubWWWDir == "" {
		return fmt.Errorf("no public www dir configured")
	}
	if err := h.requireDir(cfg.Web.PubWWWDir); err != nil {
		return fmt.Errorf("error on public www dir: %v", err)
	}
	dir := http.Dir(cfg.Web.PubWWWDir)
	h.router.NotFoundHandler = http.FileServer(dir)
	return nil
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			h.log.Crit("panic on serving HTTP", log15.Ctx{"panic": err})
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	wr := &ResponseWriter{ResponseWriter: w}
	service.SetRequestContext(r, h.ctx)
	defer service.ClearRequestContext(r)
	h.router.ServeHTTP(wr, r)
}
