// Package server provides the sesame server.
package server

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/kakeetopius/sesame/firewall"
	"github.com/yaronf/httpsign"
)

type Server struct {
	Options
	fwLock sync.RWMutex

	*mux.Router
}

type Options struct {
	Addr                string
	Port                int
	KeyFile             string
	CertFile            string
	AuthorizedClients   ClientKeys
	HTTPSEnabled        bool
	RuleDeletionEnabled bool
	RuleMaxAge          time.Duration
	FwHandler           firewall.Handler
	Logger              *log.Logger

	ruleDeletorWaitTimeout time.Duration
}

type ClientKeys map[string]ed25519.PublicKey // key is the sha256 hash encoded in hex of the PublicKey

func New(opts Options) *Server {
	if opts.AuthorizedClients == nil {
		opts.AuthorizedClients = make(ClientKeys)
	}
	if opts.Logger == nil {
		opts.Logger = log.Default()
	}

	opts.ruleDeletorWaitTimeout = 5 * time.Minute
	s := &Server{
		Options: opts,
	}

	s.addRoutes()

	return s
}

func (s *Server) Run(ctx context.Context) error {
	if s.RuleDeletionEnabled && s.RuleMaxAge <= 0 {
		return fmt.Errorf("invalid rule max age: %v", s.RuleMaxAge)
	}

	httpServer := new(http.Server{
		Addr:    net.JoinHostPort(s.Addr, strconv.Itoa(s.Port)),
		Handler: s,
	})

	if s.RuleDeletionEnabled {
		s.Logger.Println("Aged Rule Deletion Enabled.")
		go s.runRuleDeletor(ctx)
	}

	go func() {
		if s.HTTPSEnabled {
			s.Logger.Println("Server is using HTTPS")
			s.Logger.Printf("Using certificate %v and key file %v\n", s.CertFile, s.KeyFile)
			httpServer.ListenAndServeTLS(s.CertFile, s.KeyFile)
		}

		s.Logger.Println("Server is NOT using HTTPS")
		httpServer.ListenAndServe()
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := httpServer.Shutdown(shutdownCtx)
	if err != nil {
		s.Logger.Printf("Server forced to shutdown: %v\n", err)
		return nil
	}

	s.Logger.Printf("Server stopped successfully.\n")
	return nil
}

func (s *Server) runRuleDeletor(ctx context.Context) {
	ticker := time.NewTicker(s.ruleDeletorWaitTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.fwLock.Lock()
			err := s.FwHandler.DeleteAgedRules(s.RuleMaxAge)
			s.fwLock.Unlock()

			if err != nil {
				s.Logger.Printf("error deleting aged rules: %v\n", err)
			}
		}
	}
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
