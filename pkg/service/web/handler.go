package web

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/fritzpay/paymentd/pkg/paymentd/payment"

	"github.com/fritzpay/paymentd/pkg/service"
	paymentService "github.com/fritzpay/paymentd/pkg/service/payment"
	"github.com/fritzpay/paymentd/pkg/service/provider"
	"github.com/gorilla/mux"
	"gopkg.in/inconshreveable/log15.v2"
)

const (
	PaymentPath = "/payment"
)

type Handler struct {
	ctx *service.Context
	log log15.Logger

	timeout time.Duration
	router  *mux.Router

	paymentService *paymentService.Service
	templateDir    string
	keyChain       *service.Keychain

	providerService *provider.Service
}

func NewHandler(ctx *service.Context) (*Handler, error) {
	h := &Handler{
		ctx: ctx,
		log: ctx.Log().New(log15.Ctx{
			"pkg": "github.com/fritzpay/paymentd/pkg/service/web",
		}),

		router: mux.NewRouter(),
	}

	var err error
	cfg := h.ctx.Config()

	h.timeout, err = cfg.Web.Service.WriteTimeout.Duration()
	if err != nil {
		return nil, err
	}

	h.paymentService, err = paymentService.NewService(ctx)
	if err != nil {
		return nil, err
	}

	if Debug {
		w := &logIntentWorker{log: h.log}
		h.paymentService.RegisterPreIntentWorker(w)
		h.paymentService.RegisterPostIntentWorker(w)
	}

	h.providerService, err = provider.NewService(ctx)
	if err != nil {
		return nil, err
	}

	if cfg.Web.TemplateDir == "" {
		return nil, fmt.Errorf("no template dir configured")
	}
	if err := h.requireDir(cfg.Web.TemplateDir); err != nil {
		return nil, fmt.Errorf("error on template dir: %v", err)
	}
	h.templateDir = cfg.Web.TemplateDir

	err = h.registerPayment()
	if err != nil {
		h.log.Error("error registering payment", log15.Ctx{"err": err})
		return nil, err
	}

	err = h.registerPublic()
	if err != nil {
		h.log.Error("error registering www public dir", log15.Ctx{"err": err})
		return nil, err
	}

	h.log.Info("attaching provider driver endpoints...")
	err = h.providerService.AttachDrivers(h.router)
	if err != nil {
		h.log.Error("error attaching provider driver endpoints to web", log15.Ctx{"err": err})
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
	h.log.Info("registering web payment handler...")
	h.router.Handle(PaymentPath, h.ctx.RateLimitHandler(h.PaymentHandler())).Methods("GET")
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
			buf := make([]byte, 2048)
			runtime.Stack(buf, true)
			h.log.Crit("panic on serving HTTP", log15.Ctx{"panic": err, "stackTrace": string(buf)})
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	wr := &ResponseWriter{ResponseWriter: w}
	service.SetRequestContext(r, h.ctx)
	defer service.ClearRequestContext(r)
	service.TimeoutHandler(h.log.Warn, h.timeout, h.router).ServeHTTP(wr, r)
}

// debug intent worker
//
// This is a demo intent worker for the payment service
type logIntentWorker struct {
	log log15.Logger
}

func (l *logIntentWorker) PreIntent(
	p payment.Payment,
	paymentTx payment.PaymentTransaction,
	done <-chan struct{},
	res chan<- error) {
	if paymentTx.Status != payment.PaymentStatusOpen {
		return
	}
	// create a channel to receive errors from the background task
	c := make(chan error, 1)
	go func() {
		l.log.Debug("intent", log15.Ctx{
			"paymentID": p.PaymentID(),
			"status":    paymentTx.Status,
			"metadata":  p.Metadata,
		})
		if p.Metadata != nil && p.Metadata["_fBreakOpen"] != "" {
			time.Sleep(200 * time.Millisecond)
			l.log.Debug("breaking")
			c <- fmt.Errorf("breaking open")
		}
		if p.Metadata != nil && p.Metadata["_fBreakOpenLong"] != "" {
			time.Sleep(time.Second)
			c <- fmt.Errorf("break long")
		}
	}()
	// either wait for an error from the background task (this could be
	// unapproved payments, risk violations etc.)
	// or wait for the done channel to be closed and return
	select {
	case err := <-c:
		res <- err
	case <-done:
	}
}

func (l *logIntentWorker) PostIntent(p payment.Payment, paymentTx payment.PaymentTransaction) <-chan error {
	l.log.Debug("post intent", log15.Ctx{"paymentID": p.PaymentID(), "status": paymentTx.Status})
	errors := make(chan error)
	go func() {
		<-time.After(100 * time.Millisecond)
		close(errors)
	}()
	return errors
}
