package cmd

import (
	"github.com/kakeetopius/sesame/firewall"
	"github.com/kakeetopius/sesame/internal/config"
	"github.com/kakeetopius/sesame/internal/db"
	"github.com/kakeetopius/sesame/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ServerCommand() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:          "server",
		Short:        "Start the sesame server.",
		SilenceUsage: true,
		RunE:         runServer,
	}

	serverCmd.Flags().String("config", "", "Path to the config file.")
	serverCmd.Flags().StringP("addr", "a", "", "The address to use.")
	serverCmd.Flags().IntP("port", "p", 0, "The port to listen on.")
	serverCmd.Flags().Int("ssh-port", 22, "The port on which the ssh daemon is listening on.")
	return serverCmd
}

func serverConfigSetup(cmd *cobra.Command) config.ConfigSetupFunc {
	return func(v *viper.Viper) {
		err := v.BindPFlag("address", cmd.Flags().Lookup("addr"))
		cobra.CheckErr(err)
		err = v.BindPFlag("port", cmd.Flags().Lookup("port"))
		cobra.CheckErr(err)
		err = v.BindPFlag("ssh_port", cmd.Flags().Lookup("ssh-port"))
		cobra.CheckErr(err)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	config, err := config.LoadServerConfig(configPath, serverConfigSetup(cmd))
	if err != nil {
		return err
	}

	firewall, err := loadFirewall(&config)
	if err != nil {
		return err
	}

	db, err := loadDB(&config)
	if err != nil {
		return err
	}

	key, cert, err := loadSeverKeyAndCert(&config)
	if err != nil {
		return err
	}

	clientCerts, err := loadClientCerts(&config)
	if err != nil {
		return err
	}

	server := server.New(server.Options{
		Firewall:    firewall,
		DB:          db,
		Key:         key,
		Cert:        cert,
		ClientCerts: clientCerts,
		Secure:      config.Secure,
	})

	return server.Run()
}

func loadFirewall(config *config.ServerConfig) (firewall.Firewall, error) {
	return nil, nil
}

func loadDB(config *config.ServerConfig) (db.Instance, error) {
	return nil, nil
}

func loadSeverKeyAndCert(config *config.ServerConfig) (string, string, error) {
	return "", "", nil
}

func loadClientCerts(config *config.ServerConfig) ([]string, error) {
	return nil, nil
}
