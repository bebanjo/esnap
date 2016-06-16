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

	es "github.com/bebanjo/elastigo/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new repository",
	Long: `It is required to specify destination, so a new repository
will be created under this name, with a bucket named like <BUCKET><destination>
where <BUCKET> is defined in the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		var conn = es.NewConn()

		// A destination is required
		if *destination == "" {
			fmt.Fprintf(os.Stderr, "init: destination required\n")
			os.Exit(1)
		}

		log.Println("creating repository", *destination)
		if err := createRepository(conn, *destination); err != nil {
			fmt.Fprintf(os.Stderr, "create repository: error for %s %v", *destination, err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}

func createRepository(conn *es.Conn, destination string) error {
	bucket := fmt.Sprintf("%s%s", viper.Get("Bucket"), destination)
	settings := map[string]interface{}{
		"settings": map[string]interface{}{
			"bucket":                 bucket,
			"region":                 viper.Get("AZ"),
			"server_side_encryption": true,
			"protocol":               "https",
		},
		"type": "s3",
	}

	if _, err := conn.CreateSnapshotRepository(destination, nil, settings); err != nil {
		fmt.Fprintf(os.Stderr, "create repository: error for %s %v", destination, err)
		os.Exit(1)
	}

	return nil
}
