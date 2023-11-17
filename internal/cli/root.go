package cli

import (
	"log"

	"github.com/spf13/cobra"
)

var root = cobra.Command{
	Use:   "catalogd",
	Short: "list, inspect, and search for content in a catalog",
	Long:  "CLI for listing, inspecting, and searching for content provided by catalogd's Catalog resources",
}

func init() {
	root.AddCommand(&listCmd)
	root.AddCommand(&inspectCmd)
	root.AddCommand(&searchCmd)
	root.AddCommand(&versionCmd)
}

func Execute() {
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
