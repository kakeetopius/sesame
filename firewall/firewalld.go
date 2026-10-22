package firewall

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Firewalld struct {
	SSHPort int
	Zone    string
}

func NewFirewalld(sshPort int, zone string) (*Firewalld, error) {
	err := checkBinaryDependeny("firewall-cmd")
	if err != nil {
		return nil, err
	}

	f := new(Firewalld{
		SSHPort: sshPort,
		Zone:    zone,
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

func (f *Firewalld) Name() string {
	return "firewalld"
}

// AddRuleForIP enables rich rule for IP access + reloads
// example:
// firewall-cmd --permanent --zone=public --add-rich-rule='rule family="ipv4" source address="10.10.99.10/32" port protocol="tcp" port="22" accept'
func (f *Firewalld) AddRuleForIP(ipAddr string) error {
	cmd1 := exec.Command(
		`firewall-cmd`, `--permanent`,
		fmt.Sprintf("--zone=%v", f.Zone),
		fmt.Sprintf(`--add-rich-rule=rule family="ipv4" source address="%v/32" port protocol="tcp" port="%v" accept`, ipAddr, f.SSHPort),
	)
	// uncomment for debugging
	// for _, v := range cmd1.Args {
	// 	fmt.Println(v)
	// }
	output1, err1 := cmd1.CombinedOutput()
	if err1 != nil {
		return fmt.Errorf(string(output1))
	}
	fmt.Printf("rich rule added successfully for ip %v : %v", ipAddr, string(output1))

	output2, err2 := f.reload()
	if err2 != nil {
		return fmt.Errorf(string(output2))
	}
	fmt.Printf("firewalld reloaded successfully : %v", string(output2))
	return nil
}

// DeleteRuleForIP disables rich rule for IP access + reloads
func (f *Firewalld) DeleteRuleForIP(ipAddr string) error {
	cmd1 := exec.Command(
		`firewall-cmd`,
		fmt.Sprintf("--zone=%v", f.Zone),
		fmt.Sprintf(`--remove-rich-rule=rule family="ipv4" source address="%v/32" port protocol="tcp" port="%v" accept`, ipAddr, f.SSHPort),
	)
	output1, err1 := cmd1.CombinedOutput()
	if err1 != nil {
		return fmt.Errorf(string(output1))
	}
	fmt.Printf("rich rule deleted successfully for ip %v : %v", ipAddr, string(output1))

	output2, err2 := f.reload()
	if err2 != nil {
		return fmt.Errorf(string(output2))
	}
	fmt.Printf("firewalld reloaded successfully : %v", string(output2))
	return nil
}

// reload reloads firewall for setting to take effect
func (f *Firewalld) reload() ([]byte, error) {
	cmd := exec.Command("firewall-cmd", "--reload")
	output, err := cmd.CombinedOutput()
	return output, err
}

func (f *Firewalld) isActive() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "firewall-cmd", "--state")
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)

	cmd.Stderr = stderr
	cmd.Stdout = stdout

	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("firewall-cmd --state failed: %v", stderr.String())
	}

	scanner := bufio.NewScanner(stdout)
	if ok := scanner.Scan(); !ok {
		return false, fmt.Errorf("could not get firewall-cmd --state")
	}

	line := scanner.Text()
	switch strings.TrimSpace(line) {
	case "running":
		return true, nil
	case "not running":
		return false, nil
	default:
		return false, fmt.Errorf("unexpected firewall-cmd --state output: %v", stdout.String())
	}
}
