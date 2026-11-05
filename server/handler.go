package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kakeetopius/sesame/internal/ip"
)

// Index page
// GET /
func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome! You can reach sesame. You can allow an IP address using the `/ip' endpoint.\n")
}

// IPAdd for the Create action
// POST /ip
func (s *Server) IPAdd(w http.ResponseWriter, r *http.Request) {
	ipInstance := ip.Address{}
	if err := populateIPFromHandler(w, r, &ipInstance); err != nil {
		writeErrorResponse(w, http.StatusUnprocessableEntity, "Unprocessible Entity")
		return
	}

	s.fwLock.Lock()
	status, err := s.FwHandler.AddIP(ipInstance)
	s.fwLock.Unlock()
	if err != nil {
		writeErrorResponse(w, status, err.Error())
		return
	}

	writeOKResponse(w, status, "successfully added rule for ip "+ipInstance.IP)
}

// IPShow for the ip Show action
// GET /ip/{ip}
func (s *Server) IPShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ipAddr := vars["ip"]

	s.fwLock.RLock()
	ip, status, err := s.FwHandler.GetIP(ipAddr)
	s.fwLock.RUnlock()
	if err != nil {
		writeErrorResponse(w, status, err.Error())
		return
	}
	writeOKResponse(w, status, ip)
}

// ShowAllIPs shows all IPs
// GET /ip
func (s *Server) ShowAllIPs(w http.ResponseWriter, r *http.Request) {
	s.fwLock.RLock()
	ips, status, err := s.FwHandler.GetAllIPs()
	s.fwLock.RUnlock()
	if err != nil {
		writeErrorResponse(w, status, err.Error())
		return
	}
	writeOKResponse(w, status, ips)
}

// IPDelete for the ip Delete action
// DELETE /ip/{ip}
func (s *Server) IPDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ipStr := vars["ip"]

	s.fwLock.Lock()
	status, err := s.FwHandler.DeleteIP(ipStr)
	s.fwLock.Unlock()
	if err != nil {
		writeErrorResponse(w, status, err.Error())
		return
	}
	writeOKResponse(w, status, "successfully deleted rule for ip "+ipStr)
}

// Writes the response as a standard JSON response with StatusOK
func writeOKResponse(w http.ResponseWriter, status int, m any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(&JSONResponse{Data: m}); err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, "Internal Server Error")
	}
}

// Writes the error response as a Standard API JSON response with a response code
func writeErrorResponse(w http.ResponseWriter, errorCode int, errorMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(errorCode)

	json.
		NewEncoder(w).
		Encode(&JSONErrorResponse{Error: &APIError{Status: errorCode, Message: errorMsg}})
}

// Populates a ip from the params in the Handler
func populateIPFromHandler(_ http.ResponseWriter, r *http.Request, ip *ip.Address) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		return err
	}
	if err := r.Body.Close(); err != nil {
		return err
	}
	if err := json.Unmarshal(body, ip); err != nil {
		return err
	}
	return nil
}
