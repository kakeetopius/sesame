package firewall

import (
	"errors"
	"fmt"
	"os/exec"
)

type Firewall interface {
	Name() string
	AddRuleForIP(string) error
	DeleteRuleForIP(string) error
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
