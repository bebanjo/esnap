// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	es "github.com/bebanjo/elastigo/lib"
	"github.com/spf13/cobra"
)

// rotateCmd represents the rotate command
var rotateCmd = &cobra.Command{
	Use:   "rotate",
	Short: "Rotate snapshots",
	Long:  `Removes snapshots older than the given age, where default is 30 days.`,
	Run: func(cmd *cobra.Command, args []string) {
		var conn = es.NewConn()
		var now = time.Now()
		var limit = now.Add(-time.Duration(24**age) * time.Hour)
		var count int

		// Destination is required
		if *destination == "" {
			fmt.Fprintf(os.Stderr, "destination is required\n")
			os.Exit(1)
		}

		log.Println("Fetching snapshots...")
		res, err := conn.GetSnapshots(*destination, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "rotate: error %v\n", err)
			os.Exit(1)
		}

		log.Printf("Found %d snapshots on %s", len(res.Snapshots), *destination)
		log.Printf("Deleting snapshots older than %v", limit)
		for _, snapshot := range res.Snapshots {
			if int64(limit.Sub(snapshot.StartTime)) > 0 {
				if _, err := conn.DeleteSnapshot(*destination, snapshot.Snapshot); err != nil {
					fmt.Fprintf(os.Stderr, "rotate: error deleting snapshot %v\n", err)
				} else {
					log.Printf("Removed snapshot %s from %s", snapshot.Snapshot, *destination)
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
