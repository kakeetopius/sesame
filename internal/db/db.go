package db

import "github.com/kakeetopius/sesame/internal/ip"

// Instance DB interface for application
type Instance interface {
	Type() string
	GetIP(string) (*ip.IPAddress, error)
	GetAllIPs() ([]*ip.IPAddress, error)
	CheckIPExists(ipAddr string) (bool, error)
	AddIP(ip *ip.IPAddress) error
	DeleteIP(ipAddr string) (*ip.IPAddress, error)
}
