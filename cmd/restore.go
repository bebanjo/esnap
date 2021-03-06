package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	es "github.com/bebanjo/elastigo/lib"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a snapshot",
	Long: `You are required to set an origin, destination, and snapshot name.
By default, it will fetch the given snapshot from the origin repository, creating
new indices out of the ones from the snapshot, and make a swap of the alias, removing
the old indices. If you use the fresh option, all indices and alias will be restored,
without a swap.`,
	Run: func(cmd *cobra.Command, args []string) {
		var conn = es.NewConn()
		var date = time.Now().Format("20060102150405")

		// Origin, destination and snapshot names are required
		if *originRestore == "" || *destination == "" || *snapshot == "" {
			fmt.Fprintf(os.Stderr, "origin, destination and snapshot are required\n")
			os.Exit(1)
		}

		// fresh restore
		if *fresh {
			log.Println("applying fresh restore")
			if err := freshRestore(conn, *originRestore, *destination, *snapshot, date); err != nil {
				fmt.Fprintf(os.Stderr, "fresh restore: error %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}

		// restore without recreating aliases
		if err := restore(conn, *originRestore, *destination, *snapshot, date); err != nil {
			fmt.Fprintf(os.Stderr, "restore: error %v\n", err)
			os.Exit(1)
		}

		// iterate aliases to do the swap
		suffix := fmt.Sprintf("%s%s", date, *snapshot)
		aliasesInfo := conn.GetCatAliasInfo(fmt.Sprintf("%s*", *destination))
		for _, aliasInfo := range aliasesInfo {
			indicesInfo := conn.GetCatIndexInfo(fmt.Sprintf("%s*", aliasInfo.Name))
			indicesNames := indicesNames(indicesInfo)
			var indicesNamesToDelete []string
			var disableDeletion bool

			// iterate new created indices matching the alias pattern
			for _, indexName := range indicesNames {
				if indexName == aliasInfo.Index || !strings.HasSuffix(indexName, suffix) {
					indicesNamesToDelete = append(indicesNamesToDelete, indexName)
					continue
				}

				// add alias when new index is green
				if err := addAliasPolling(conn, aliasInfo.Name, indexName); err != nil {
					fmt.Fprintf(os.Stderr, "add alias: error with alias %s and index %s %v\n", aliasInfo.Name, indexName, err)
					disableDeletion = true
					continue
				}
			}

			// do not delete old indices if an alias to a new index failed to be created
			if disableDeletion {
				log.Println("restore finished without deletions, see errors above")
				os.Exit(0)
			}

			// delete old indices
			for _, indexNameToDelete := range indicesNamesToDelete {
				if _, err := conn.DeleteIndex(indexNameToDelete); err != nil {
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

func freshRestore(conn *es.Conn, origin, destination, snapshotName, date string) error {
	query := map[string]interface{}{
		"ignore_unavailable":   true,
		"include_global_state": false,
		"rename_pattern":       fmt.Sprintf("%s_(.+)_\\d+(_.*)?", origin),
		"rename_replacement":   fmt.Sprintf("%s_$1_%s%s", destination, date, snapshotName),
	}

	_, err := conn.RestoreSnapshot(origin, snapshotName, nil, query)
	return err
}

func restore(conn *es.Conn, origin, destination, snapshotName, date string) error {
	query := map[string]interface{}{
		"ignore_unavailable":   "true",
		"include_global_state": false,
		"include_aliases":      false,
		"rename_pattern":       fmt.Sprintf("%s_(.+)_\\d+(_.*)?", origin),
		"rename_replacement":   fmt.Sprintf("%s_$1_%s%s", destination, date, snapshotName),
	}

	_, err := conn.RestoreSnapshot(origin, snapshotName, nil, query)
	return err
}

func addAliasPolling(conn *es.Conn, aliasName, indexName string) error {
	var state string
	log.Printf("index %s is in status... ", indexName)
	for state != "green" {
		indexInfo := conn.GetCatIndexInfo(indexName)
		if len(indexInfo) < 1 {
			break
		}

		state = indexInfo[0].Health
		if state == "green" {
			log.Printf("index %s is in status... %s ", indexName, state)
			log.Println("Adding alias", aliasName, "to index", indexName)
			if _, err := conn.AddAlias(indexName, aliasName); err != nil {
				return err
			}

			return nil
		}
		log.Printf("index %s is in status... %s ", indexName, state)

		time.Sleep(3 * time.Second)
	}

	return nil
}
