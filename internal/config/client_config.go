package config

type ClientConfig struct {
	Address     string `mapstructure:"address"`
	Port        int    `mapstructure:"port"`
	CertFile    string `mapstructure:"cert_file"`
	PrivKeyFile string `mapstructure:"private_key_file"`
}

func LoadClientConfig(path string, setUpFuncs ...ConfigSetupFunc) (ClientConfig, error) {
	return ClientConfig{}, nil
}
