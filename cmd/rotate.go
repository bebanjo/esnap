package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate snapshots",
	Long:  `Removes snapshots older than the given age, where default is 30 days.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := mustClient()
		now := time.Now()
		limit := now.Add(-time.Duration(24**age) * time.Hour)
		count := 0

		if *destination == "" {
			fmt.Fprintf(os.Stderr, "destination is required\n")
			os.Exit(1)
		}

		log.Println("Fetching snapshots...")
		res, err := client.GetSnapshots(*destination)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rotate: error %v\n", err)
			os.Exit(1)
		}

		log.Printf("Found %d snapshots on %s", len(res), *destination)
		log.Printf("Deleting snapshots older than %v", limit)
		for _, snapshot := range res {
			if int64(limit.Sub(snapshot.StartTime)) > 0 {
				if err := client.DeleteSnapshot(*destination, snapshot.Name); err != nil {
					fmt.Fprintf(os.Stderr, "rotate: error deleting snapshot %v\n", err)
				} else {
					log.Printf("Removed snapshot %s from %s", snapshot.Name, *destination)
					count++
				}
			}
		}
		log.Printf("%d snapshots on %s were rotated", count, *destination)
	},
}

func init() {
	RootCmd.AddCommand(rotateCmd)

	age = rotateCmd.PersistentFlags().IntP("age", "a", 30, "Maximun age in days to keep snapshots")
}
