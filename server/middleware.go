package server

import (
	"log"
	"net/http"

	"github.com/yaronf/httpsign"
)

func (s *Server) validateMiddleware(next http.Handler) http.Handler {
	config := httpsign.NewHandlerConfig().
		SetFetchVerifier(s.fetchVerifier).
		SetReqNotVerified(s.onVerifyFail)

	return httpsign.WrapHandler(next, *config)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (s *Server) fetchVerifier(r *http.Request) (string, *httpsign.Verifier) {
	sigName := SignatureName()

	details, err := httpsign.RequestDetails(sigName, r)
	if err != nil || details.KeyID == nil {
		return sigName, nil
	}

	key, ok := s.AuthorizedClients[*details.KeyID]
	if !ok {
		return sigName, nil
	}

	mustBeSigned := MustBeSignedFields(r.Method)

	v, err := httpsign.NewEd25519Verifier(key,
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
