package cmd

import (
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "sesame",
		Short:        "A rest application to update firewall rules on a linux server.",
		SilenceUsage: true,
	}

	rootCmd.AddCommand(
		ServerCommand(),
		ClientCommand(),
	)

	return rootCmd
}
