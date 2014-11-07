package server

import (
	"errors"
	"fmt"
	"github.com/facebookgo/grace"
	"github.com/fritzpay/paymentd/pkg/config"
	"golang.org/x/net/context"
	"gopkg.in/inconshreveable/log15.v2"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	serverWaitTimeout = 10 * time.Second
)

var Wait sync.WaitGroup

// Server is a  paymentd server
type Server struct {
	ctx    context.Context
	log    log15.Logger
	Cancel context.CancelFunc

	httpServers []*http.Server

	// graceful restart
	// net.Listeners for graceful restart
	listeners []grace.Listener
	// errors while serving
	errors chan error
	// shutdown chan, will be closed when cleaned up
	shutdown chan struct{}
}

// NewServer creates a new paymentd server for the given config
func NewServer(ctx context.Context) *Server {
	srv := &Server{
		httpServers: make([]*http.Server, 0, 3),

		shutdown: make(chan struct{}),
	}
	srv.ctx = ctx
	if log, ok := srv.ctx.Value("log").(log15.Logger); ok {
		srv.log = log
	} else {
		srv.log = log15.New()
		srv.log.SetHandler(log15.StderrHandler)
	}
	srv.log = srv.log.New(log15.Ctx{"pkg": "github.com/fritzpay/paymentd/pkg/server"})
	return srv
}

// RegisterService adds a service to the server
// It will serve the HTTP with the given service
func (s *Server) RegisterService(cfg config.ServiceConfig, handler http.Handler) error {
	srv := &http.Server{
		Addr:           cfg.Address,
		Handler:        handler,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}
	var err error
	srv.ReadTimeout, err = cfg.ReadTimeout.Duration()
	if err != nil {
		return fmt.Errorf("error parsing duration for server %s: %v", cfg.Address, err)
	}
	srv.WriteTimeout, err = cfg.WriteTimeout.Duration()
	if err != nil {
		return fmt.Errorf("error parsing duration for server %s: %v", cfg.Address, err)
	}
	s.httpServers = append(s.httpServers, srv)
	return nil
}

// Serve starts serving
func (s *Server) Serve() error {
	if len(s.httpServers) == 0 {
		return errors.New("no services registered")
	}
	inherited, err := s.listen()
	if err != nil {
		return err
	}
	pid := os.Getpid()
	ppid := os.Getppid()
	if inherited {
		if ppid == 1 {
			for _, l := range s.listeners {
				s.log.Info("server listening on init activate", log15.Ctx{
					"address": l.Addr().String(),
					"PID":     pid,
				})
			}
		} else {
			const msg = "graceful handoff"
			for _, l := range s.listeners {
				s.log.Info(msg, log15.Ctx{
					"address": l.Addr().String(),
					"newPID":  pid,
					"oldPID":  ppid,
				})
			}
		}
	} else {
		for _, l := range s.listeners {
			s.log.Info("server listening", log15.Ctx{
				"address": l.Addr().String(),
				"PID":     pid,
			})
		}
	}

	s.serveHTTP()

	// inherited? not init activated? close parent
	if inherited && os.Getppid() != 1 {
		if err := grace.CloseParent(); err != nil {
			s.log.Crit("error closing parent process", log15.Ctx{
				"pid":  pid,
				"ppid": ppid,
			})
			return fmt.Errorf("error closing parent: %v", err)
		}
	}

	err = s.wait()

	<-s.shutdown

	s.log.Info("exiting. graceful handoff complete.", log15.Ctx{
		"pid": pid,
	})

	return err
}

// create listeners
// will return true if inheriting, false if not
func (s *Server) listen() (bool, error) {
	var err error
	s.errors = make(chan error, len(s.httpServers))
	// try to inherit listeners from parent
	s.listeners, err = grace.Inherit()
	if err == nil {
		// previous listeners are present
		if len(s.listeners) != len(s.httpServers) {
			return true, fmt.Errorf("listener handoff mismatch. got %d listeners and %d servers", len(s.listeners), len(s.httpServers))
		}
		return true, nil
	} else if err == grace.ErrNotInheriting {
		// new listeners
		if s.listeners, err = s.newListeners(); err != nil {
			return false, err
		}
		return false, nil
	}
	return false, fmt.Errorf("error on graceful handoff: %v", err)
}

func (s *Server) newGraceListener(addrStr string) (grace.Listener, error) {
	addr, err := net.ResolveTCPAddr("tcp", addrStr)
	if err != nil {
		return nil, fmt.Errorf("error resolving address %s: %v", addr, err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error listening on address %s: %v", addr, err)
	}
	return grace.NewListener(l), nil
}

// create new listeners
func (s *Server) newListeners() ([]grace.Listener, error) {
	listeners := make([]grace.Listener, len(s.httpServers))
	var err error
	for i := 0; i < len(s.httpServers); i++ {
		listeners[i], err = s.newGraceListener(s.httpServers[i].Addr)
		if err != nil {
			return nil, err
		}
	}
	return listeners, nil
}

// serve all HTTP servers without blocking
func (s *Server) serveHTTP() {
	for i, l := range s.listeners {
		go func(i int, l grace.Listener) {
			server := s.httpServers[i]
			err := server.Serve(l)
			if err != nil && err != grace.ErrAlreadyClosed {
				s.errors <- fmt.Errorf("error serving HTTP %s: %v", server.Addr, err)
			}
		}(i, l)
	}
}

// the final blocking, wait for server to stop serving
func (s *Server) wait() error {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM)
	go func() {
		<-sigs
		s.Shutdown()
	}()

	waiterr := make(chan error)
	go func() {
		waiterr <- grace.Wait(s.listeners)
	}()
	select {
	case err := <-waiterr:
		return err
	case err := <-s.errors:
		return err
	}
}
