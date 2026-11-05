package firewall

import (
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/kakeetopius/sesame/internal/db"
	"github.com/kakeetopius/sesame/internal/ip"
)

type Handler struct {
	HandlerOptions
}

type HandlerOptions struct {
	DB       db.Instance
	Firewall Firewall
	Logger   *log.Logger
}

var (
	ErrIPExists    = errors.New("ip already exists")
	ErrIPNotExists = errors.New("ip does not exist")
)

func NewHandler(opts HandlerOptions) Handler {
	return Handler{HandlerOptions: opts}
}

func (h Handler) AddIP(ipAddr ip.Address) (int, error) {
	ipInstance, err := ip.New(ipAddr.IP) // New will validate the ip and populate some other ip fields not given in JSON.
	if err != nil {
		return http.StatusUnprocessableEntity, err
	}
	ipExists, err := h.DB.CheckIPExists(ipInstance.IP)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if ipExists {
		return http.StatusBadRequest, ErrIPExists
	}

	ipInstance.Added = time.Now()

	env := os.Getenv("env")
	if env != "local" {
		err = h.Firewall.AddRuleForIP(ipInstance)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	err = h.DB.AddIP(ipInstance)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusCreated, nil
}

func (h Handler) DeleteIP(ipAddr string) (int, error) {
	ipInstance, err := ip.New(ipAddr)
	if err != nil {
		return http.StatusUnprocessableEntity, err
	}
	ipExists, err := h.DB.CheckIPExists(ipInstance.IP)
	if err != nil {
		return http.StatusBadRequest, err
	}

	if !ipExists {
		return http.StatusNotFound, ErrIPNotExists
	}

	env := os.Getenv("env")
	if env != "local" {
		err = h.Firewall.DeleteRuleForIP(ipInstance)
		if err != nil {
			return http.StatusInternalServerError, err
		}
	}

	err = h.DB.DeleteIP(ipInstance.IP)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

func (h Handler) GetIP(ipAddr string) (ip.Address, int, error) {
	ipInstance, err := h.DB.GetIP(ipAddr)
	if err != nil {
		return ip.Address{}, http.StatusBadRequest, err
	}
	return ipInstance, http.StatusOK, nil
}

func (h Handler) GetAllIPs() ([]ip.Address, int, error) {
	ips, err := h.DB.GetAllIPs()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return ips, http.StatusOK, nil
}

func (h Handler) DeleteAgedRules(maxAge time.Duration) error {
	ips, err := h.DB.GetAllIPs()
	if err != nil {
		return err
	}
	if len(ips) == 0 {
		return nil
	}

	var errs []error
	now := time.Now()
	for _, ip := range ips {
		age := now.Sub(ip.Added)
		if age < maxAge {
			continue
		}

		h.Logger.Printf("deleting aged rule for %v, added at %v (%v ago)\n", ip, ip.Added, age)
		err = h.Firewall.DeleteRuleForIP(ip)
		if err != nil {
			return err
		}

		err = h.DB.DeleteIP(ip.IP)
		if err != nil {
			return err
		}
	}

	return errors.Join(errs...)
}
