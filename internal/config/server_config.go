// Package config handles the app's configuration.
package config

import (
	"fmt"
	"net/netip"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Addr               string          `mapstructure:"address"`
	Port               int             `mapstructure:"port"`
	SSHPort            int             `mapstructure:"ssh_port"`
	Insecure           bool            `mapstructure:"insecure"` // use http
	DBFile             string          `mapstructure:"db"`
	CertFile           string          `mapstructure:"cert_file"`
	KeyFile            string          `mapstructure:"key_file"`
	AuthorizedCertsDir string          `mapstructure:"authorized_certs_dir"`
	Firewalld          FirewalldConfig `mapstructure:"firewalld"`
	Ufw                UfwConfig       `mapstructure:"ufw"`
}

type FirewalldConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Zone    string `mapstructure:"zone"`
}

type UfwConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type ConfigSetupFunc func(*viper.Viper)

func LoadServerConfig(path string, setUpFuncs ...ConfigSetupFunc) (ServerConfig, error) {
	serverConf := ServerConfig{}

	viperConf, err := initConfig(path, "sesame", "/etc/sesame/")
	if err != nil {
		return ServerConfig{}, err
	}

	for _, f := range setUpFuncs {
		f(viperConf)
	}

	err = viperConf.Unmarshal(&serverConf)
	if err != nil {
		return ServerConfig{}, err
	}

	serverConf.setDefaults()

	err = serverConf.Validate()
	if err != nil {
		return ServerConfig{}, err
	}

	return serverConf, nil
}

func (c *ServerConfig) Validate() error {
	_, err := netip.ParseAddr(c.Addr)
	if err != nil {
		return fmt.Errorf("invalid listening IP address: ", c.Addr)
	}
	if c.Port <= 0 || c.Port >= 65536 {
		return fmt.Errorf("invalid listening port: ", c.Port)
	}
	if c.SSHPort <= 0 || c.SSHPort >= 65536 {
		return fmt.Errorf("invalid SSH port: ", c.SSHPort)
	}

	if !c.Insecure {
		if c.CertFile == "" {
			return fmt.Errorf("https enabled but no path to server certificate file given.")
		}
		if c.KeyFile == "" {
			return fmt.Errorf("https enabled but no path to server private key file given.")
		}
	}

	if c.AuthorizedCertsDir == "" {
		return fmt.Errorf("no path to the clients' certificates directory given.")
	}

	if c.Ufw.Enabled && c.Firewalld.Enabled {
		return fmt.Errorf("cannot enable more than one firewall backend.")
	}

	if !c.Ufw.Enabled && !c.Firewalld.Enabled {
		return fmt.Errorf("no firewall backends are enabled.")
	}

	return nil
}

func (c *ServerConfig) setDefaults() {
	if c.Addr == "" {
		c.Addr = "0.0.0.0"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.SSHPort == 0 {
		c.SSHPort = 22
	}
	if c.DBFile == "" {
		c.DBFile = "/var/sesame/sesame.db"
	}
	if c.AuthorizedCertsDir == "" {
		c.AuthorizedCertsDir = "/etc/sesame/authorized_certs"
	}

	if c.Firewalld.Enabled && c.Firewalld.Zone == "" {
		c.Firewalld.Zone = "public"
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cfgFile string, defaultName string, defaultPath string) (*viper.Viper, error) {
	appConfig := viper.New()

	if cfgFile != "" {
		appConfig.SetConfigFile(cfgFile)
	} else {
		if defaultName == "" {
			return nil, fmt.Errorf("could not initialise configuration: no config file given")
		}
		appConfig.SetConfigName(defaultName)
		if defaultPath != "" {
			appConfig.AddConfigPath(defaultPath)
		}
	}

	appConfig.AutomaticEnv() // read in environment variables that match

	if err := appConfig.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file %v: %w", viper.ConfigFileUsed(), err)
	}

	return appConfig, nil
}
