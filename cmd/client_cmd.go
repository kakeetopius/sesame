package cmd

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/kakeetopius/sesame/internal/config"
	"github.com/kakeetopius/sesame/internal/ip"
	"github.com/kakeetopius/sesame/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yaronf/httpsign"
)

func ClientCommand() *cobra.Command {
	clientCmd := &cobra.Command{
		Use:          "client",
		Short:        "Make requests to a sesame server.",
		SilenceUsage: true,
	}

	clientCmd.PersistentFlags().String("config", "", "Path to the client config file.")
	clientCmd.PersistentFlags().StringP("addr", "a", "", "The address to connect to.")
	clientCmd.PersistentFlags().IntP("port", "p", 0, "The port to connect to.")
	clientCmd.PersistentFlags().String("cert-file", "", "Path to the client certificate.")
	clientCmd.PersistentFlags().String("key-file", "", "Path to the client private key file.")
	clientCmd.PersistentFlags().Bool("skip-verify", false, "Skip verifying the server's certificate if connecting with TLS.")

	clientCmd.AddCommand(
		ClientAllowCommand(),
		ClientDenyCommand(),
	)

	return clientCmd
}

func ClientAllowCommand() *cobra.Command {
	allowCmd := &cobra.Command{
		Use:          "allow ip",
		Short:        "Allow SSH access on the server for an ip",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         runClientAllow,
	}

	return allowCmd
}

func ClientDenyCommand() *cobra.Command {
	denyCmd := &cobra.Command{
		Use:          "deny ip",
		Short:        "Deny SSH access on the server for an ip",
		SilenceUsage: true,
		Args:         cobra.ExactArgs(1),
		RunE:         runClientDeny,
	}

	return denyCmd
}

func clientConfigSetupFunc(cmd *cobra.Command) config.ConfigSetupFunc {
	return func(v *viper.Viper) {
		err := v.BindPFlag("address", cmd.Flags().Lookup("addr"))
		cobra.CheckErr(err)
		err = v.BindPFlag("port", cmd.Flags().Lookup("port"))
		cobra.CheckErr(err)
		err = v.BindPFlag("cert_file", cmd.Flags().Lookup("cert-file"))
		cobra.CheckErr(err)
		err = v.BindPFlag("private_key_file", cmd.Flags().Lookup("key-file"))
		cobra.CheckErr(err)
	}
}

func runClientAllow(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no ip given")
	}

	skipVerify, err := cmd.Flags().GetBool("skip-verify")
	if err != nil {
		return err
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	clientConfig, err := config.LoadClientConfig(configPath, clientConfigSetupFunc(cmd))
	if err != nil {
		return err
	}

	key, keyID, err := loadClientKeyAndID(&clientConfig)
	if err != nil {
		return err
	}

	signer, err := httpsign.NewEd25519Signer(
		key,
		httpsign.
			NewSignConfig().
			SetKeyID(keyID),
		*server.MustBeSignedFields("POST"),
	)
	if err != nil {
		return err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipVerify,
			},
		},
	}

	client := httpsign.NewClient(
		*httpClient,
		httpsign.
			NewClientConfig().
			SetSignatureName(server.SignatureName()).
			SetSigner(signer),
	)

	ip, err := ip.New(args[0])
	if err != nil {
		return err
	}

	body, err := json.Marshal(ip)
	if err != nil {
		return fmt.Errorf("json error: %w", err)
	}

	url, err := getServerURL(clientConfig.Address, clientConfig.Port, "/m1/ip")
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body)

	return nil
}

func runClientDeny(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("no ip given")
	}

	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	skipVerify, err := cmd.Flags().GetBool("skip-verify")
	if err != nil {
		return err
	}

	clientConfig, err := config.LoadClientConfig(configPath, clientConfigSetupFunc(cmd))
	if err != nil {
		return err
	}

	key, keyID, err := loadClientKeyAndID(&clientConfig)
	if err != nil {
		return err
	}

	signer, err := httpsign.NewEd25519Signer(
		key,
		httpsign.
			NewSignConfig().
			SetKeyID(keyID),
		*server.MustBeSignedFields("DELETE"),
	)
	if err != nil {
		return err
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipVerify,
			},
		},
	}

	client := httpsign.NewClient(
		*httpClient,
		httpsign.
			NewClientConfig().
			SetSignatureName(server.SignatureName()).
			SetSigner(signer),
	)

	ip, err := ip.New(args[0])
	if err != nil {
		return err
	}

	url, err := getServerURL(clientConfig.Address, clientConfig.Port, "/m1/ip/"+ip.IP)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	io.Copy(os.Stdout, resp.Body)

	return nil
}

// path must start with a '/'
func getServerURL(address string, port int, path string) (string, error) {
	var addr string
	if strings.HasPrefix(address, "https://") || strings.HasPrefix(address, "http://") {
		addr = fmt.Sprintf("%v:%v%v", address, port, path)
	} else {
		addr = fmt.Sprintf("//%v:%v%v", address, port, path) // the '//' prevent url.Parse from treating the 'address' as scheme
	}

	url, err := url.Parse(addr)
	if err != nil {
		return "", fmt.Errorf("invalid address: %v", address)
	}

	if url.Scheme == "" {
		url.Scheme = "http"
	}

	return url.String(), nil
}

func loadClientKeyAndID(config *config.ClientConfig) (ed25519.PrivateKey, string, error) {
	privKey, err := os.ReadFile(config.PrivKeyFile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read client private key file: %w", err)
	}

	block, _ := pem.Decode(privKey)
	if block == nil {
		return nil, "", fmt.Errorf("private key: %v. No pem block found", config.PrivKeyFile)
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse client private key: %w -> Note that only ed25519 keys are allowed", err)
	}

	ed25519Key, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, "", fmt.Errorf("private key %v. Only ed25519 private keys allowed", config.PrivKeyFile)
	}

	certFile, err := os.ReadFile(config.CertFile)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read client certificate: %w", err)
	}

	block, _ = pem.Decode(certFile)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, "", fmt.Errorf("%v: no CERTIFICATE block found", config.CertFile)
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, "", fmt.Errorf("failed to parse client certificate: %w", err)
	}

	pubKey, ok := cert.PublicKey.(ed25519.PublicKey)
	if !ok {
		return nil, "", fmt.Errorf("certificate %v. Only certificates with an ed25519 public keys are allowed", config.CertFile)
	}

	pubKeyHash := sha256.Sum256(pubKey)

	return ed25519Key, hex.EncodeToString(pubKeyHash[:]), nil
}
