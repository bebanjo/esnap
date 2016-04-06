package cmd

import (
	es "github.com/bebanjo/esnap/_vendor/src/github.com/mattbaird/elastigo/lib"
	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup unused indices",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var conn = es.NewConn()
		indicesInfo := conn.GetCatIndexInfo("")
		_ = indicesNames(indicesInfo)
	},
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
}
