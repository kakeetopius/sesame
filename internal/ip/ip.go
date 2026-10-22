// Package ip contains the definition of an IP address.
package ip

import (
	"fmt"
	"net/netip"
	"strings"
)

type IPAddress struct {
	IP string `json:"ip"`
}

func (i *IPAddress) Verify() error {
	if strings.Contains(i.IP, "/") {
		prefix, err := netip.ParsePrefix(i.IP)
		if err != nil {
			return fmt.Errorf("invalid IP address: ", i.IP)
		}
		if prefix.Addr().Is4() && prefix.Bits() != 32 {
			return fmt.Errorf("only host IP addresses ie /32 are allowed")
		}
		if prefix.Addr().Is6() && prefix.Bits() != 128 {
			return fmt.Errorf("only host IP addresses ie /128 are allowed")
		}

		i.IP = strings.Split(i.IP, "/")[0]

		return nil
	}

	_, err := netip.ParseAddr(i.IP)
	if err != nil {
		return fmt.Errorf("invalid IP address: ", i.IP)
	}

	return nil
}
