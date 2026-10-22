package cmd

import "github.com/spf13/cobra"

func ServerCommand() *cobra.Command {
	serverCmd := &cobra.Command{
		Use:          "server",
		Short:        "Start the sesame server.",
		SilenceUsage: true,
	}

	return serverCmd
}
