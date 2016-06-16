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

// takeCmd represents the snapshot take command
var takeCmd = &cobra.Command{
	Use:   "take",
	Short: "Take a snapshot",
	Long: `You are required to set a destination. It will create a snapshot
on the destination repository. If repository does not exist, you can create
it with the provided flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		var conn = es.NewConn()
		var date = time.Now().Format("20060102150405")
		var state = "STARTING"
		var query interface{}

		// A destination is required
		if *destination == "" {
			fmt.Fprintf(os.Stderr, "take: destination required\n")
			os.Exit(1)
		}

		// Create repository if --create-repository flag is enabled
		if *createRepositoryTake {
			log.Println("creating repository", *destination)
			if err := createRepository(conn, *destination); err != nil {
				fmt.Fprintf(os.Stderr, "create repository: error for %s %v", *destination, err)
				os.Exit(1)
			}
		}

		// Select only destinationTake-related indices if --all flag is not used
		if !*allIndices {
			indicesInfo := conn.GetCatIndexInfo(fmt.Sprintf("%s*", *destination))
			indicesNamesString := strings.Join(indicesNames(indicesInfo), ",")
			query = map[string]interface{}{"indices": indicesNamesString}
		}

		// Take Snapshot
		_, err := conn.TakeSnapshot(*destination, date, nil, query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "take: error %v\n", err)
			os.Exit(1)
		}

		// Poll for Snapshot status until it is done
		log.Println("waiting for snapshot", date, "to be ready...", state)
		for state != "SUCCESS" {
			snapshots, err := conn.GetSnapshotByName(*destination, date, nil)
			if err != nil {
				fmt.Fprintf(os.Stderr, "take: error getting snapshot %s %v\n", *destination, err)
				os.Exit(1)
			}

			if len(snapshots.Snapshots) < 1 {
				break
			}

			state = snapshots.Snapshots[0].State
			log.Println("waiting for snapshot", date, "to be ready...", state)
			time.Sleep(5 * time.Second)
		}

	},
}

func init() {
	RootCmd.AddCommand(takeCmd)

	createRepositoryTake = takeCmd.PersistentFlags().BoolP("create-repository", "r", false, "Create repository")
	allIndices = takeCmd.PersistentFlags().BoolP("all", "a", false,
		"Take snapshot of all indices. Otherwise, only those matching the destination")
}

func indicesNames(catIndexInfo []es.CatIndexInfo) []string {
	var names []string
	for _, cii := range catIndexInfo {
		names = append(names, cii.Name)
	}
	return names
}
