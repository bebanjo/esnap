package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	es "github.com/bebanjo/esnap/vendor/src/github.com/mattbaird/elastigo/lib"
	"github.com/spf13/cobra"
)

// takeCmd represents the snapshot take command
var takeCmd = &cobra.Command{
	Use:   "take",
	Short: "Take a snapshot",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var conn = es.NewConn()
		var date = time.Now().Format("20060102150405")
		var state = "STARTING"
		var query interface{}

		// A destinationTake is required
		if *destinationTake == "" {
			fmt.Fprintf(os.Stderr, "take: destination required\n")
			os.Exit(1)
		}

		// Create repository if --create-repository flag is enabled
		if *createRepository {
			fmt.Println("creating repository", destinationTake)
			repositoryType := map[string]interface{}{"type": "s3"}
			settings := map[string]interface{}{
				"bucket":                 fmt.Sprintf("bebanjo-elasticsearch-snapshots-%s", *destinationTake),
				"region":                 "eu-west-1",
				"server_side_encryption": true,
				"protocol":               "https",
			}
			if _, err := conn.CreateSnapshotRepository(*destinationTake, repositoryType, settings); err != nil {
				fmt.Fprintf(os.Stderr, "create repository: error for %s %v", *destinationTake, err)
				os.Exit(1)
			}
		}

		// Select only destinationTake-related indices if --all flag is not used
		if !*allIndices {
			indicesInfo := conn.GetCatIndexInfo(fmt.Sprintf("%s*", *destinationTake))
			indicesNamesString := strings.Join(indicesNames(indicesInfo), ",")
			query = map[string]interface{}{"indices": indicesNamesString}
		}

		// Take Snapshot
		_, err := conn.TakeSnapshot(*destinationTake, date, nil, query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "take: error %v\n", err)
			os.Exit(1)
		}

		// Poll for Snapshot status until it is done
		fmt.Println("waiting for snapshot", date, "to be ready...", state)
		for state != "SUCCESS" {
			snapshots, err := conn.GetSnapshotByName(*destinationTake, date, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "take: error getting snapshot %s %v\n", *destinationTake, err)
				os.Exit(1)
			}

			if len(snapshots.Snapshots) < 1 {
				break
			}

			state = snapshots.Snapshots[0].State
			fmt.Println("waiting for snapshot", date, "to be ready...", state)
			time.Sleep(5 * time.Second)
		}

	},
}

func init() {
	RootCmd.AddCommand(takeCmd)

	createRepository = takeCmd.PersistentFlags().BoolP("create-repository", "r", false, "Create repository")
	destinationTake = takeCmd.PersistentFlags().StringP("destination-take", "d", "", "Destination of the snapshot")
	allIndices = takeCmd.PersistentFlags().BoolP("all", "a", false,
		"Take snapshot of all indices. Otherwise, only those matching the destination-take")
}

func indicesNames(catIndexInfo []es.CatIndexInfo) []string {
	var names []string
	for _, cii := range catIndexInfo {
		names = append(names, cii.Name)
	}
	return names
}
