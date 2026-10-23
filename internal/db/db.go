// Package db provides the different database implementations that can persist and retrieve IP addresses
package db

import "github.com/kakeetopius/sesame/internal/ip"

// Instance DB interface for application
type Instance interface {
	Type() string
	GetIP(string) (*ip.Address, error)
	GetAllIPs() ([]*ip.Address, error)
	CheckIPExists(ipAddr string) (bool, error)
	AddIP(ip *ip.Address) error
	DeleteIP(ipAddr string) error
}
