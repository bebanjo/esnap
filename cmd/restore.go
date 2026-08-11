package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	esclient "github.com/bebanjo/esnap/internal/es"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a snapshot",
	Long: `You are required to set an origin, destination, and snapshot name.
By default, it will fetch the given snapshot from the origin repository, creating
new indices out of the ones from the snapshot, and make a swap of the alias, removing
the old indices. If you use the fresh option, all indices and alias will be restored,
without a swap.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustClient()
		date := time.Now().Format("20060102150405")

		if *originRestore == "" || *destination == "" || *snapshot == "" {
			fmt.Fprintf(os.Stderr, "origin, destination and snapshot are required\n")
			os.Exit(1)
		}

		if *fresh {
			log.Println("applying fresh restore")
			if err := freshRestore(client, *originRestore, *destination, *snapshot, date); err != nil {
				fmt.Fprintf(os.Stderr, "fresh restore: error %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		if err := restore(client, *originRestore, *destination, *snapshot, date); err != nil {
			fmt.Fprintf(os.Stderr, "restore: error %v\n", err)
			os.Exit(1)
		}

		suffix := date
		aliasesInfo, err := client.GetAliases(fmt.Sprintf("%s*", *destination))
		if err != nil {
			fmt.Fprintf(os.Stderr, "restore: error fetching aliases %v\n", err)
			os.Exit(1)
		}

		for _, aliasInfo := range aliasesInfo {
			indicesInfo, err := client.GetIndices(fmt.Sprintf("%s*", aliasInfo.Name))
			if err != nil {
				fmt.Fprintf(os.Stderr, "restore: error fetching indices for alias %s %v\n", aliasInfo.Name, err)
				os.Exit(1)
			}
			indicesNames := indicesNames(indicesInfo)
			indicesNamesToDelete, indicesNamesToAlias := partitionRestoredIndices(indicesNames, aliasInfo.Index, suffix)
			var disableDeletion bool

			for _, indexName := range indicesNamesToAlias {
				if err := addAliasPolling(client, aliasInfo.Name, indexName); err != nil {
					fmt.Fprintf(os.Stderr, "add alias: error with alias %s and index %s %v\n", aliasInfo.Name, indexName, err)
					disableDeletion = true
					continue
				}
			}

			if disableDeletion {
				log.Println("restore finished without deletions, see errors above")
				os.Exit(0)
			}

			for _, indexNameToDelete := range indicesNamesToDelete {
				if err := client.DeleteIndex(indexNameToDelete); err != nil {
					fmt.Fprintf(os.Stderr, "delete index: error with index %s %v\n", indexNameToDelete, err)
				}
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(restoreCmd)

	originRestore = restoreCmd.PersistentFlags().StringP("origin", "o", "",
		"Origin of the snapshot to restore")
	snapshot = restoreCmd.PersistentFlags().StringP("snapshot", "s", "",
		"Name of the snapshot to restore")
	fresh = restoreCmd.PersistentFlags().BoolP("fresh", "f", false,
		"Do a full, fresh restore of all data")
}

func freshRestore(client *esclient.Client, origin, destination, snapshotName, date string) error {
	return client.RestoreSnapshot(origin, snapshotName, esclient.RestoreOptions{
		IgnoreUnavailable:  true,
		IncludeGlobalState: false,
		IncludeAliases:     true,
		RenamePattern:      fmt.Sprintf("%s_(.+)_\\d+(_.*)?", origin),
		RenameReplacement:  fmt.Sprintf("%s_$1_%s", destination, date),
	})
}

func restore(client *esclient.Client, origin, destination, snapshotName, date string) error {
	return client.RestoreSnapshot(origin, snapshotName, esclient.RestoreOptions{
		IgnoreUnavailable:  true,
		IncludeGlobalState: false,
		IncludeAliases:     false,
		RenamePattern:      fmt.Sprintf("%s_(.+)_\\d+(_.*)?", origin),
		RenameReplacement:  fmt.Sprintf("%s_$1_%s", destination, date),
	})
}

// partitionRestoredIndices splits indicesNames into indices to delete (stale
// indices from a previous restore, or the currently aliased index) and
// indices to alias (freshly restored indices from this run, identified by
// ending in suffix).
func partitionRestoredIndices(indicesNames []string, currentAliasedIndex, suffix string) (toDelete, toAlias []string) {
	for _, indexName := range indicesNames {
		if indexName == currentAliasedIndex || !strings.HasSuffix(indexName, suffix) {
			toDelete = append(toDelete, indexName)
			continue
		}
		toAlias = append(toAlias, indexName)
	}
	return toDelete, toAlias
}

func addAliasPolling(client *esclient.Client, aliasName, indexName string) error {
	state := ""
	log.Printf("index %s is in status... ", indexName)
	for state != "green" {
		indexInfo, err := client.GetIndices(indexName)
		if err != nil {
			return err
		}
		if len(indexInfo) < 1 {
			break
		}

		state = indexInfo[0].Health
		if state == "green" {
			log.Printf("index %s is in status... %s ", indexName, state)
			log.Println("Adding alias", aliasName, "to index", indexName)
			if err := client.AddAlias(indexName, aliasName); err != nil {
				return err
			}
			return nil
		}
		log.Printf("index %s is in status... %s ", indexName, state)
		time.Sleep(3 * time.Second)
	}

	return nil
}
