// Package firewall provides the different firewall backends that can be used to dynamically add firewall rules to allow or block SSH for a particular IP address.
package firewall

import (
	"errors"
	"fmt"
	"os/exec"

	"github.com/kakeetopius/sesame/internal/ip"
)

type Firewall interface {
	Name() string
	AddRuleForIP(ip.Address) error
	DeleteRuleForIP(ip.Address) error
}

type ErrFirewallNotActive struct {
	Name string
}

type ErrFirewallNotFound struct {
	Name string
}

func (e ErrFirewallNotActive) Error() string {
	return fmt.Sprintf("firewall '%s' is not active", e.Name)
}

func (e ErrFirewallNotFound) Error() string {
	return fmt.Sprintf("firewall command '%s' not found.", e.Name)
}

func checkBinaryDependeny(name string) error {
	_, err := exec.LookPath(name)
	if err != nil {
		if execErr, ok := errors.AsType[*exec.Error](err); ok {
			if errors.Is(execErr, exec.ErrNotFound) {
				return ErrFirewallNotFound{Name: name}
			}
		}
		return err
	}
	return nil
}
