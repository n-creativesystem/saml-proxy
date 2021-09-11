package cmd

import (
	"fmt"

	"github.com/n-creativesystem/saml-proxy/version"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Long:  "Print the version number",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("version: %s, Revision: %s\n", version.Version, version.Revision)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
