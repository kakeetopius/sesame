package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/kakeetopius/sesame/internal/ip"
)

// Index page
// GET /
func (s *Server) Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome! You can reach sesame.\n")
}

// IPAdd for the Create action
// POST /ip
func (s *Server) IPAdd(w http.ResponseWriter, r *http.Request) {
	ipInstance := ip.Address{}
	if err := populateIPFromHandler(w, r, &ipInstance); err != nil {
		writeErrorResponse(w, http.StatusUnprocessableEntity, "Unprocessible Entity")
		return
	}

	ipInstance, err := ip.New(ipInstance.IP) // New will validate the ip and populate some other fields not given in JSON.
	if err != nil {
		writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	ipExists, err := s.DB.CheckIPExists(ipInstance.IP)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ipExists {
		writeErrorResponse(w, http.StatusBadRequest, "ip already exists")
		return
	}

	env := os.Getenv("env") // WHY? - maybe testing - TODO
	if env != "local" {
		err = s.Firewall.AddRuleForIP(ipInstance)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to add ip , err : %v", err.Error()))
			return
		}
	}

	err = s.DB.AddIP(&ipInstance)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeOKResponse(w, "successfully added rule for ip "+ipInstance.IP)
}

// IPShow for the ip Show action
// GET /ip/{ip}
func (s *Server) IPShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ipAddr := vars["ip"]
	ip, err := s.DB.GetIP(ipAddr)
	if err != nil {
		// No IP found
		writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}
	writeOKResponse(w, ip)
}

// ShowAllIPs shows all IPs
// GET /ip
func (s *Server) ShowAllIPs(w http.ResponseWriter, r *http.Request) {
	ips, err := s.DB.GetAllIPs()
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeOKResponse(w, ips)
}

// IPDelete for the ip Delete action
// DELETE /ip/{ip}
func (s *Server) IPDelete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ipStr := vars["ip"]

	ipAddr, _ := ip.New(ipStr)

	ipExists, err := s.DB.CheckIPExists(ipStr)
	if err != nil {
		writeErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ipExists {
		writeErrorResponse(w, http.StatusNotFound, "ip does not exist")
		return
	}

	env := os.Getenv("env")
	if env != "local" {
		err = s.Firewall.DeleteRuleForIP(ipAddr)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("failed to delete ip, err : %v", err.Error()))
			return
		}
	}

	err = s.DB.DeleteIP(ipStr)
	if err != nil {
		// IP could not be deleted
		writeErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}
	writeOKResponse(w, "successfully deleted rule for ip "+ipStr)
}

// Writes the response as a standard JSON response with StatusOK
func writeOKResponse(w http.ResponseWriter, m any) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

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
