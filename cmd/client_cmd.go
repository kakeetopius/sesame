package cmd

import "github.com/spf13/cobra"

func ClientCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:          "client",
		Short:        "Make requests to a sesame server.",
		SilenceUsage: true,
	}

	return rootCmd
}
