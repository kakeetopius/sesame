package server

import (
	"github.com/gorilla/mux"
	"github.com/kakeetopius/sesame/firewall"
)

type Server struct {
	Firewall        firewall.Firewall
	KeyFile         string
	CertFile        string
	ClientCertsFile string

	*mux.Router
}

func New(firewall firewall.Firewall) *Server {
	return &Server{
		Router:   newRouter(),
		Firewall: firewall,
	}
}
