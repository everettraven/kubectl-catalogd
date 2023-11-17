package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "dev"
var catalogd_version = "unknown"

var versionCmd = cobra.Command{
	Use:   "version",
	Short: "version of the kubectl-catalogd plugin",
	Long:  "Prints the version of the kubectl-catalogd plugin",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("kubectl-catalogd version:", version, ", catalogd version:", catalogd_version)
	},
}
