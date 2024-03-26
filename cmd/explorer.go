package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ignite/gex/pkg/view"
)

const defaultHost = "http://localhost:26657"

// NewExplorer creates a new explorer command.
func NewExplorer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gex [host]",
		Short: "Run gex explorer",
		Long:  "Gex is a cosmos explorer for terminals",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			host := defaultHost
			if args[0] != "" {
				host = args[0]
			}
			fmt.Println(host)

			v, err := view.DrawView()
			if err != nil {
				return err
			}

			return v.Run(cmd.Context())
		},
	}

	return cmd
}
