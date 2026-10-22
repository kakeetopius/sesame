// Package config handles the app's configuration.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Addr           string          `mapstructure:"address"`
	Port           int             `mapstructure:"port"`
	CertFile       string          `mapstructure:"cert_file"`
	KeyFile        string          `mapstructure:"key_file"`
	ClientCertsDir string          `mapstructure:"client_certs_dir"`
	Firewalld      FirewalldConfig `mapstructure:"firewalld"`
	Ufw            UfwConfig       `mapstructure:"ufw"`
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
	return ServerConfig{}, nil
}

// func Load()
// func verify()
// func setDefaults()

// initConfig reads in config file and ENV variables if set.
func initConfig(cfgFile string, defaultName string, defaultPath string) (*viper.Viper, error) {
	appConfig := viper.New()

	if cfgFile != "" {
		// Use config file from the flag.
		appConfig.SetConfigFile(cfgFile)
	} else {
		appConfig.SetConfigName(defaultName)
		appConfig.AddConfigPath(defaultPath)
	}

	appConfig.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := appConfig.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// if file not found no need to error.
			return appConfig, nil
		}
		return nil, fmt.Errorf("error reading config file %v: %w", viper.ConfigFileUsed(), err)
	}

	return appConfig, nil
}
