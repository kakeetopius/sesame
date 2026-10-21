package firewall

import (
	"fmt"
)

type Firewall interface {
	AddRuleForIP(string) (string, error)
	DeleteRuleForIP(string) (string, error)
	Check() error
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
