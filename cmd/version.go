package cmd

import (
	"github.com/spf13/cobra"

	"github.com/ignite/gex/version"
)

// NewVersion creates a new version command to show the Ignite CLI version.
func NewVersion() *cobra.Command {
	c := &cobra.Command{
		Use:   "version",
		Short: "Print the current build information",
		Run: func(cmd *cobra.Command, _ []string) {
			cmd.Println(version.Long(cmd.Context()))
		},
	}
	return c
}
