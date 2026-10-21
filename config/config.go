// Package config handles the apps configuration.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type ServerConfig struct {
	Addr            string
	Port            int
	CertFile        string
	KeyFile         string
	ClientCertsFile string
	Firewalld       FirewalldConfig
	Ufw             UfwConfig
}

type FirewalldConfig struct {
	Enabled bool
	Zone    string
}

type UfwConfig struct {
	Enabled bool
}

type ConfigSetupFunc func(*viper.Viper)

func LoadServerConfig(path string, setUpFuncs ...ConfigSetupFunc) (ServerConfig, error) {
	return ServerConfig{}, nil
}

// func Load()
// func verify()
// func setDefaults()

// initConfig reads in config file and ENV variables if set.
func initConfig(cfgFile string) (*viper.Viper, error) {
	appConfig := viper.New()

	if cfgFile != "" {
		// Use config file from the flag.
		appConfig.SetConfigFile(cfgFile)
	} else {
		appConfig.SetConfigName("firewalld-rest")
		appConfig.AddConfigPath("/etc/firewalld-rest")
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
