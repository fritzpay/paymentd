// +build go1.3

package server

import (
	"gopkg.in/inconshreveable/log15.v2"
	"time"
)

// Shutdown starts the server's shutdown mode
//
// It will cancel all server child contexts, disable Keepalive on all servers
func (s *Server) Shutdown() {
	s.log.Warn("server going into shutdown mode")
	// SetKeepAlivesEnabled introduced in Go 1.3
	for _, srv := range s.httpServers {
		srv.SetKeepAlivesEnabled(false)
	}
	if s.Cancel != nil {
		s.Cancel()
	}
	waited := make(chan struct{})
	go func() {
		Wait.Wait()
		close(waited)
	}()
	select {
	case <-waited:
	case <-time.After(serverWaitTimeout):
		s.log.Warn("server exiting after wait timeout", log15.Ctx{"waitTimeout": serverWaitTimeout})
	}
	close(s.shutdown)
}
