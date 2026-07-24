package cmd

import (
	"fmt"
	"log"
	"os"

	esclient "github.com/bebanjo/esnap/internal/es"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Cleanup unused indices",
	Long: `It will find all indices that are not pointed by an alias.
Handle with care in case this is an expected scenario!`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustClient()

		indicesInfo, err := client.GetIndices("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "cleanup: error fetching indices %v\n", err)
			os.Exit(1)
		}
		aliasesInfo, err := client.GetAliases("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "cleanup: error fetching aliases %v\n", err)
			os.Exit(1)
		}

		indicesNamesToDelete := indicesNamesToDelete(indicesInfo, aliasesInfo)
		for _, indexNameToDelete := range indicesNamesToDelete {
			if err := client.DeleteIndex(indexNameToDelete); err != nil {
				fmt.Fprintf(os.Stderr, "delete index: error with index %s %v\n", indexNameToDelete, err)
			}
			log.Printf("deleting index %s... OK", indexNameToDelete)
		}
	},
}

func init() {
	RootCmd.AddCommand(cleanupCmd)
}

func indicesNamesToDelete(indicesInfo []esclient.Index, aliasesInfo []esclient.Alias) []string {
	toDelete := make([]string, 0, len(indicesInfo))

	for _, indexInfo := range indicesInfo {
		found := false
		for _, aliasInfo := range aliasesInfo {
			if indexInfo.Name == aliasInfo.Index {
				found = true
				break
			}
		}

		if !found {
			toDelete = append(toDelete, indexInfo.Name)
		}
	}

	return toDelete
}
