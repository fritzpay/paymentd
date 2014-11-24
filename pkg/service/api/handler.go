package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fritzpay/paymentd/pkg/service"
	"github.com/fritzpay/paymentd/pkg/service/api/v1"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

// Handler is the (HTTP) API Handler
type Handler struct {
	ctx *service.Context
	log log15.Logger

	timeout time.Duration
	mux     *mux.Router
}

// NewHandler creates a new API Handler
func NewHandler(ctx *service.Context) (*Handler, error) {
	h := &Handler{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/api",
		}),

		mux: mux.NewRouter(),
	}

	var err error
	// Serve Admin GUI if active and path provided
	cfg := h.ctx.Config()

	h.timeout, err = cfg.API.Timeout.Duration()
	if err != nil {
		return nil, err
	}

	adminGUIPubWWWDir := cfg.API.AdminGUIPubWWWDir
	if cfg.API.ServeAdmin && len(adminGUIPubWWWDir) > 0 {
		err = h.requireDir(adminGUIPubWWWDir)
		if err != nil {
			h.log.Error("error reading admin gui www public dir", log15.Ctx{"err": err})
			return nil, err
		}
		err = h.registerPublic()
		if err != nil {
			h.log.Error("error registering admin gui www public dir", log15.Ctx{"err": err})
			return nil, err
		}
	}

	h.log.Info("registering API service v1...")
	v1.NewService(h.ctx, h.mux)
	v1.Log = h.log.New(log15.Ctx{
		"pkg": "github.com/fritzpay/paymentd/pkg/service/api/v1",
	})

	return h, nil
}

// ServeHTTP implements the http.Handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			h.log.Crit("panic on serving HTTP", log15.Ctx{"panic": err})
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	service.SetRequestContext(r, h.ctx)
	defer service.ClearRequestContext(r)
	h.mux.ServeHTTP(w, r)
	// service.TimeoutHandler(h.log.Warn, h.timeout, h.mux).ServeHTTP(w, r)
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

func (h *Handler) registerPublic() error {
	h.log.Info("registering www public admin gui directory...")
	cfg := h.ctx.Config()
	if cfg.API.AdminGUIPubWWWDir == "" {
		return fmt.Errorf("no public admin gui www dir configured")
	}
	if err := h.requireDir(cfg.API.AdminGUIPubWWWDir); err != nil {
		return fmt.Errorf("error on public admin gui www dir: %v", err)
	}
	dir := http.Dir(cfg.API.AdminGUIPubWWWDir)
	h.mux.NotFoundHandler = http.FileServer(dir)
	return nil
}
