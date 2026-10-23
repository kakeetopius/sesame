package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type ClientConfig struct {
	Address     string `mapstructure:"address"`
	Port        int    `mapstructure:"port"`
	CertFile    string `mapstructure:"cert_file"`
	PrivKeyFile string `mapstructure:"private_key_file"`
}

func LoadClientConfig(path string, setUpFuncs ...ConfigSetupFunc) (ClientConfig, error) {
	clientConf := ClientConfig{}

	var defaultPath string

	if path == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			return ClientConfig{}, fmt.Errorf("error getting user's configuration directory: %w", err)
		}
		defaultPath = filepath.Join(configDir, "sesame")
	}

	viperConf, err := initConfig(path, "sesame", defaultPath)
	if err != nil {
		if _, ok := errors.AsType[viper.ConfigFileNotFoundError](err); !ok {
			return ClientConfig{}, err
		}
		viperConf = viper.New() // create a new viper config not bound to any config file.
	}

	for _, f := range setUpFuncs {
		f(viperConf)
	}

	err = viperConf.Unmarshal(&clientConf)
	if err != nil {
		return ClientConfig{}, err
	}

	clientConf.setDefaults()

	err = clientConf.Validate()
	if err != nil {
		return ClientConfig{}, err
	}

	return clientConf, nil
}

func (c *ClientConfig) Validate() error {
	if c.Address == "" {
		return fmt.Errorf("no server IP given to connect to")
	}
	if c.Port <= 0 || c.Port >= 65536 {
		return fmt.Errorf("invalid server port: %v", c.Port)
	}

	if c.CertFile == "" {
		return fmt.Errorf("no certificate file path given")
	}

	if c.PrivKeyFile == "" {
		return fmt.Errorf("no private key file given")
	}

	return nil
}

func (c *ClientConfig) setDefaults() {
	if c.Port == 0 {
		c.Port = 8080
	}
}
