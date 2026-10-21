package server

import (
	"github.com/gorilla/mux"
	"github.com/prashantgupta24/firewalld-rest/firewall"
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
