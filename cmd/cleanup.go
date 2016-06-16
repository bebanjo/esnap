package cmd

import (
	"fmt"
	"log"
	"os"

	es "github.com/bebanjo/elastigo/lib"
	"github.com/spf13/cobra"
)

// cleanupCmd represents the cleanup command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup unused indices",
	Long: `It will find all indices that are not pointed by an alias.
Handle with care in case this is an expected scenario!`,
	Run: func(cmd *cobra.Command, args []string) {
		var conn = es.NewConn()

		indicesInfo := conn.GetCatIndexInfo("")
		aliasesInfo := conn.GetCatAliasInfo("")

		indicesNamesToDelete := indicesNamesToDelete(indicesInfo, aliasesInfo)
		for _, indexNameToDelete := range indicesNamesToDelete {
			_, err := conn.DeleteIndex(indexNameToDelete)
			if err != nil {
				fmt.Fprintf(os.Stderr, "delete index: error with index %s %v\n", indexNameToDelete, err)
			}
			log.Printf("deleting index %s... OK", indexNameToDelete)
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
}

func indicesNamesToDelete(indicesInfo []es.CatIndexInfo, aliasesInfo []es.CatAliasInfo) []string {
	var toDelete []string
	var found bool

	for _, indexInfo := range indicesInfo {
		for _, aliasInfo := range aliasesInfo {
			if indexInfo.Name == aliasInfo.Index {
				found = true
				break
			}
		}

		if !found {
			toDelete = append(toDelete, indexInfo.Name)
		}
		found = false
	}

	return toDelete
}
