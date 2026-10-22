package cmd

import "github.com/spf13/cobra"

func ClientCommand() *cobra.Command {
	clientCmd := &cobra.Command{
		Use:          "client",
		Short:        "Make requests to a sesame server.",
		SilenceUsage: true,
	}

	clientCmd.Flags().String("config", "", "Path to the config file.")
	return clientCmd
}
