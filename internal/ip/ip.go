// Package ip contains the definition of an IP address.
package ip

import (
	"fmt"
	"net/netip"
	"strings"
)

type Address struct {
	IP     string `json:"ip"`
	Family string `json:"-"`
	Bits   int    `json:"-"`
}

func New(s string) (Address, error) {
	var ip Address

	if strings.Contains(s, "/") {
		prefix, err := netip.ParsePrefix(s)
		if err != nil {
			return ip, fmt.Errorf("invalid IP address: ", s)
		}

		if prefix.Addr().Is4() {
			if prefix.Bits() != 32 {
				return ip, fmt.Errorf("only host IP addresses ie /32 are allowed")
			}
			ip.Family = "ipv4"
			ip.Bits = 32
		}
		if prefix.Addr().Is6() {
			if prefix.Bits() != 128 {
				return ip, fmt.Errorf("only host IP addresses ie /128 are allowed")
			}
			ip.Family = "ipv6"
			ip.Bits = 128
		}

		s = strings.Split(s, "/")[0]
		ip.IP = s
	} else {
		addr, err := netip.ParseAddr(s)
		if err != nil {
			return ip, fmt.Errorf("invalid IP address: ", s)
		}

		ip.IP = s
		if addr.Is4() {
			ip.Family = "ipv4"
			ip.Bits = 32
		} else if addr.Is6() {
			ip.Family = "ipv6"
			ip.Bits = 128
		}

	}

	return ip, nil
}
