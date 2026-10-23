package cmd

import (
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

	serverCmd.Flags().String("config", "", "Path to the server config file.")
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
	if os.Geteuid() != 0 {
		return fmt.Errorf("the sesame server requires root priviledges to run")
	}

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

	clientCerts, err := loadAuthorizedCerts(&config)
	if err != nil {
		return err
	}

	serverOpts := server.Options{
		Addr:              config.Addr,
		Port:              config.Port,
		KeyFile:           config.KeyFile,
		CertFile:          config.CertFile,
		Firewall:          firewall,
		DB:                db,
		AuthorizedClients: clientCerts,
		Insecure:          config.Insecure,
		Logger:            log.Default(),
	}

	server := server.New(serverOpts)

	return server.Run()
}

func loadFirewall(config *config.ServerConfig) (firewall.Firewall, error) {
	switch {
	case config.Firewalld.Enabled:
		firewalld, err := firewall.NewFirewalld(config.SSHPort, config.Firewalld.Zone)
		if err != nil {
			return nil, err
		}
		log.Println("Using firewall backend 'firewalld'")
		return firewalld, nil
	case config.Ufw.Enabled:
		ufw, err := firewall.NewUfw(config.SSHPort)
		if err != nil {
			return nil, err
		}
		log.Println("Using firewall backend 'ufw'")
		return ufw, nil
	default:
		return nil, fmt.Errorf("no firewall is enabled")
	}
}

func loadDB(config *config.ServerConfig) (db.Instance, error) {
	log.Println("Using database file ", config.DBFile)
	return db.GetFileTypeInstance(config.DBFile)
}

func loadAuthorizedCerts(config *config.ServerConfig) (server.ClientKeys, error) {
	certPool := make(server.ClientKeys)
	log.Println("Using client authorized directory ", config.AuthorizedCertsDir)

	authorizedDir, err := os.ReadDir(config.AuthorizedCertsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read authorized certificates directory: %w", err)
	}

	if len(authorizedDir) == 0 {
		log.Println("No authorized certificates found in ", config.AuthorizedCertsDir)
		return certPool, nil
	}

	for _, file := range authorizedDir {
		if file.IsDir() {
			continue
		}
		certFile, err := os.ReadFile(filepath.Join(config.AuthorizedCertsDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read certificate file %v: %w", file.Name(), err)
		}

		block, _ := pem.Decode(certFile)
		if block == nil || block.Type != "CERTIFICATE" {
			return nil, fmt.Errorf("client certificate %v: no CERTIFICATE block found", file.Name())
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate file %v: %w", file.Name(), err)
		}

		pubKey, ok := cert.PublicKey.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("certificate file: %v. Only certificates with ed25519 public keys are allowed", file.Name())
		}

		sha := sha256.Sum256(pubKey)

		shaHex := hex.EncodeToString(sha[:])

		certPool[shaHex] = pubKey
	}

	log.Printf("Found %v certificates in the authorized clients directory %v\n", len(certPool), config.AuthorizedCertsDir)

	return certPool, nil
}
