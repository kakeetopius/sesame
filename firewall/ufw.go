package firewall

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/kakeetopius/sesame/internal/ip"
)

type Ufw struct {
	SSHPort int
}

func (f *Ufw) Name() string {
	return "ufw"
}

func NewUfw(sshPort int) (*Ufw, error) {
	err := checkBinaryDependeny("ufw")
	if err != nil {
		return nil, err
	}

	f := new(Ufw{
		SSHPort: sshPort,
	})

	active, err := f.isActive()
	if err != nil {
		return nil, err
	}
	if !active {
		return nil, ErrFirewallNotActive{Name: f.Name()}
	}

	return f, nil
}

// AddRuleForIP enables rule for IP access + reloads
// example:
func (f *Ufw) AddRuleForIP(ipAddr ip.Address) error {
	cmd1 := exec.Command(
		"ufw", "allow",
		"from", ipAddr.IP,
		"to", "any",
		"port", strconv.Itoa(f.SSHPort),
		"proto", "tcp",
	)
	// uncomment for debugging
	// for _, v := range cmd1.Args {
	// 	fmt.Println(v)
	// }
	output1, err1 := cmd1.CombinedOutput()
	if err1 != nil {
		return fmt.Errorf(string(output1))
	}
	fmt.Printf("rule added successfully for ip %v : %v", ipAddr.IP, string(output1))

	output2, err2 := f.reload()
	if err2 != nil {
		return fmt.Errorf(string(output2))
	}
	fmt.Printf("ufw reloaded successfully : %v", string(output2))
	return nil
}

// DeleteRuleForIP disables rule for IP access + reloads
func (f *Ufw) DeleteRuleForIP(ipAddr ip.Address) error {
	cmd1 := exec.Command(
		"ufw", "delete", "allow",
		"from", ipAddr.IP,
		"to", "any",
		"port", strconv.Itoa(f.SSHPort),
		"proto", "tcp",
	)
	output1, err1 := cmd1.CombinedOutput()
	if err1 != nil {
		return fmt.Errorf(string(output1))
	}
	fmt.Printf("rule deleted successfully for ip %v : %v", ipAddr.IP, string(output1))

	output2, err2 := f.reload()
	if err2 != nil {
		return fmt.Errorf(string(output2))
	}
	fmt.Printf("ufw reloaded successfully : %v", string(output2))
	return nil
}

// reload reloads firewall for setting to take effect
func (f *Ufw) reload() ([]byte, error) {
	cmd := exec.Command("ufw", "reload")
	output, err := cmd.CombinedOutput()
	return output, err
}

func (f *Ufw) isActive() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ufw", "status")
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd.Stderr = stderr
	cmd.Stdout = stdout

	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("error command ufw status failed: %v", stderr.String())
	}

	scanner := bufio.NewScanner(stdout)
	if ok := scanner.Scan(); !ok {
		return false, fmt.Errorf("could not get ufw status")
	}

	line := scanner.Text() // format is 'Status: inactive' or 'Status: active'
	parts := strings.Split(line, ":")
	if len(parts) != 2 {
		return false, fmt.Errorf("unexpected ufw status output: %v", stdout.String())
	}
	switch strings.TrimSpace(parts[1]) {
	case "active":
		return true, nil
	case "inactive":
		return false, nil
	default:
		return false, fmt.Errorf("unexpected ufw status output: %v", stdout.String())
	}
}
