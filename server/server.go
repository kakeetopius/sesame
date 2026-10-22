package server

import (
	"crypto/ed25519"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kakeetopius/sesame/firewall"
	"github.com/kakeetopius/sesame/internal/db"
	"github.com/yaronf/httpsign"
)

type Server struct {
	Options
	*mux.Router
}

type Options struct {
	Addr              string
	Port              int
	Firewall          firewall.Firewall
	DB                db.Instance
	KeyFile           string
	CertFile          string
	AuthorizedClients ClientKeys
	Insecure          bool
}

type ClientKeys map[string]ed25519.PublicKey // key is the sha256 hash encoded in hex of the PublicKey

func New(opts Options) *Server {
	s := &Server{
		Options: opts,
	}

	s.addRoutes()

	return s
}

func (s *Server) Run() error {
	if s.Insecure {
		return http.ListenAndServe(fmt.Sprintf("%v:%v", s.Addr, s.Port), s)
	}
	return http.ListenAndServeTLS(fmt.Sprintf("%v:%v", s.Addr, s.Port), s.CertFile, s.KeyFile, s)
}

func MustBeSignedFields(method string) *httpsign.Fields {
	fields := httpsign.Headers("@method", "@authority", "@path")

	switch strings.ToUpper(method) {
	case "POST":
		return fields.AddHeader("content-digest")
	default:
		return &fields
	}
}

func SignatureName() string {
	return "magic-word"
}
