package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	esclient "github.com/bebanjo/esnap/internal/es"
	"github.com/spf13/cobra"
)

var takeCmd = &cobra.Command{
	Use:   "take",
	Short: "Take a snapshot",
	Long: `You are required to set a destination. It will create a snapshot
on the destination repository. If repository does not exist, you can create
it with the provided flag.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustClient()
		date := time.Now().Format("20060102150405")
		state := "STARTING"
		var indicesNamesToTake []string
		var indicesFilterRule string

		if *destination == "" {
			fmt.Fprintf(os.Stderr, "take: destination required\n")
			os.Exit(1)
		}

		if *createRepositoryTake {
			log.Println("creating repository", *destination)
			if err := createRepository(client, *destination); err != nil {
				fmt.Fprintf(os.Stderr, "create repository: error for %s %v", *destination, err)
				os.Exit(1)
			}
		}

		if !*allIndices {
			indicesFilterRule = fmt.Sprintf("%s*", *destination)
		}

		indicesInfo, err := client.GetIndices(indicesFilterRule)
		if err != nil {
			fmt.Fprintf(os.Stderr, "take: error fetching indices %v\n", err)
			os.Exit(1)
		}
		indicesNamesToTake = indicesNames(indicesInfo)

		if *aliased {
			aliasesInfo, err := client.GetAliases(indicesFilterRule)
			if err != nil {
				fmt.Fprintf(os.Stderr, "take: error fetching aliases %v\n", err)
				os.Exit(1)
			}
			indicesNamesToTake = aliasedIndicesNames(aliasesInfo, indicesNamesToTake)
		}

		log.Println("Taking snapshot of indices:", joinNames(indicesNamesToTake))
		if err := client.TakeSnapshot(*destination, date, indicesNamesToTake); err != nil {
			fmt.Fprintf(os.Stderr, "take: error %v\n", err)
			os.Exit(1)
		}

		log.Println("waiting for snapshot", date, "to be ready...", state)
		for state != "SUCCESS" {
			snapshot, err := client.GetSnapshot(*destination, date)
			if err != nil {
				fmt.Fprintf(os.Stderr, "take: error getting snapshot %s, id %s %v\n", *destination, date, err)
				os.Exit(1)
			}
			if snapshot == nil {
				break
			}

			state = snapshot.State
			log.Println("waiting for snapshot", date, "to be ready...", state)
			if state == "FAILED" || state == "PARTIAL" {
				fmt.Fprintf(os.Stderr, "take: snapshot %s finished with state %s\n", date, state)
				os.Exit(1)
			}
			time.Sleep(5 * time.Second)
		}
	},
}

func init() {
	RootCmd.AddCommand(takeCmd)

	createRepositoryTake = takeCmd.PersistentFlags().BoolP("create-repository", "r", false, "Create repository")
	allIndices = takeCmd.PersistentFlags().BoolP("all", "a", false,
		"Take snapshot of all indices. Otherwise, only those matching the destination")
	aliased = takeCmd.PersistentFlags().BoolP("aliased", "", false,
		"Take snapshot of indices with associated aliases only")
}

func indicesNames(indicesInfo []esclient.Index) []string {
	names := make([]string, 0, len(indicesInfo))
	for _, indexInfo := range indicesInfo {
		names = append(names, indexInfo.Name)
	}
	return names
}

func aliasedIndicesNames(aliasesInfo []esclient.Alias, indicesNames []string) []string {
	names := make([]string, 0, len(indicesNames))
	for _, aliasInfo := range aliasesInfo {
		for _, indexName := range indicesNames {
			if aliasInfo.Index == indexName {
				names = append(names, indexName)
				break
			}
		}
	}
	return names
}

func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	joined := names[0]
	for _, name := range names[1:] {
		joined += "," + name
	}
	return joined
}
