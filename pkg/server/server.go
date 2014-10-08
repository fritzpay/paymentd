package server

import (
	"github.com/fritzpay/paymentd/pkg/config"
)

// Server is a  paymentd server
type Server struct {
}

// NewServer creates a new paymentd server for the given config
func NewServer(cfg config.Config) (*Server, error) {
	srv := &Server{}
	return srv, nil
}
