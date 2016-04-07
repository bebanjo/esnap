package cmd

import (
	"fmt"

	es "github.com/bebanjo/esnap/vendor/src/github.com/mattbaird/elastigo/lib"
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
		aliasesInfo := conn.GetCatAliasInfo("")

		indicesToRemove := indicesToRemove(indicesInfo, aliasesInfo)
		for _, indexToRemove := range indicesToRemove {
			fmt.Printf("Deleting index %s... ", indexToRemove)
			_, err := conn.DeleteIndex(indexToRemove)
			if err != nil {
				fmt.Printf("ERROR: %v\n", err)
			}
			fmt.Println("OK")
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
}

func indicesToRemove(indicesInfo []es.CatIndexInfo, aliasesInfo []es.CatAliasInfo) []string {
	var toRemove []string
	var found bool

	for _, indexInfo := range indicesInfo {
		for _, aliasInfo := range aliasesInfo {
			if indexInfo.Name == aliasInfo.Index {
				found = true
				break
			}
		}

		if !found {
			toRemove = append(toRemove, indexInfo.Name)
		}
		found = false
	}

	return toRemove
}
