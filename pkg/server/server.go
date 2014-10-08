package server

import (
	"errors"
	"github.com/fritzpay/paymentd/pkg/config"
	"net/http"
)

// Server is a  paymentd server
type Server struct {
	httpServers []*http.Server
}

// NewServer creates a new paymentd server for the given config
func NewServer() *Server {
	srv := &Server{
		httpServers: make([]*http.Server, 0, 3),
	}
	return srv
}

// RegisterService adds a service to the server
// It will serve the HTTP with the given service
func (s *Server) RegisterService(cfg config.ServiceConfig, handler http.Handler) {
	srv := &http.Server{
		Addr:           cfg.Address,
		Handler:        handler,
		ReadTimeout:    cfg.ReadTimeout,
		WriteTimeout:   cfg.WriteTimeout,
		MaxHeaderBytes: cfg.MaxHeaderBytes,
	}
	s.httpServers = append(s.httpServers, srv)
}

// Serve starts serving
func (s *Server) Serve() error {
	if len(s.httpServers) == 0 {
		return errors.New("no services registered")
	}
	errors := make(chan error)
	for _, srv := range s.httpServers {
		go func(srv *http.Server) {
			err := srv.ListenAndServe()
			if err != nil {
				errors <- err
			}
		}(srv)
	}
	select {
	case err := <-errors:
		return err
	}
}
