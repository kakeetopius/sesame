package server

import (
	"github.com/gorilla/mux"
	"github.com/kakeetopius/sesame/firewall"
	"github.com/kakeetopius/sesame/internal/db"
)

type Server struct {
	Firewall       firewall.Firewall
	KeyFile        string
	CertFile       string
	ClientCertsDir string

	DB db.Instance
	*mux.Router
}

func New(firewall firewall.Firewall) *Server {
	s := &Server{
		Firewall: firewall,
	}

	s.addRoutes()

	return s
}
