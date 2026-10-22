package server

import (
	"github.com/gorilla/mux"
	"github.com/kakeetopius/sesame/firewall"
	"github.com/kakeetopius/sesame/internal/db"
)

type Server struct {
	Options
	*mux.Router
}

type Options struct {
	Firewall    firewall.Firewall
	DB          db.Instance
	Key         string
	Cert        string
	ClientCerts []string
	Insecure    bool
}

func New(opts Options) *Server {
	s := &Server{
		Options: opts,
	}

	s.addRoutes()

	return s
}

func (s *Server) Run() error {
	return nil
}
