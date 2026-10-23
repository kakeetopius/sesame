package server

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/yaronf/httpsign"
)

func (s *Server) validateMiddleware(next http.Handler) http.Handler {
	config := httpsign.
		NewHandlerConfig().
		SetFetchVerifier(s.fetchVerifier).
		SetReqNotVerified(s.onVerifyFail)

	return httpsign.WrapHandler(next, *config)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.Logger.Printf("Request from %s to %s\n", r.RemoteAddr, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func (s *Server) recoveryMiddleWare(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				s.Logger.Printf("PANIC RECOVERED: %v\n%s", err, debug.Stack())

				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (s *Server) fetchVerifier(r *http.Request) (string, *httpsign.Verifier) {
	sigName := SignatureName()

	details, err := httpsign.RequestDetails(sigName, r)
	if err != nil || details.KeyID == nil {
		s.Logger.Println("got a signature without a key id from ", r.RemoteAddr)
		return sigName, nil
	}

	if details.Alg != "ed25519" { // only ed25519 allowed.
		s.Logger.Println("got a signature that didn't use ed25519 from ", r.RemoteAddr)
		return sigName, nil
	}

	key, ok := s.AuthorizedClients[*details.KeyID]
	if !ok {
		return sigName, nil
	}

	mustBeSigned := MustBeSignedFields(r.Method)

	v, err := httpsign.NewEd25519Verifier(
		key,
		httpsign.NewVerifyConfig().SetKeyID(*details.KeyID),
		*mustBeSigned,
	)
	if err != nil {
		return sigName, nil
	}

	return sigName, v
}

func (s *Server) onVerifyFail(w http.ResponseWriter, r *http.Request, l *log.Logger, err error) {
	if l != nil {
		l.Printf("signature verification failed from %s: %v", r.RemoteAddr, err)
	}
	http.Error(w, "invalid signature", http.StatusUnauthorized)
}
