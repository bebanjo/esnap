package cmd

import (
	"fmt"
	"log"
	"os"

	esclient "github.com/bebanjo/esnap/internal/es"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Creates a new repository",
	Long: `It is required to specify destination, so a new repository
will be created under this name, with a bucket named like <BUCKET><destination>
where <BUCKET> is defined in the configuration.`,
	Run: func(cmd *cobra.Command, args []string) {
		if *destination == "" {
			fmt.Fprintf(os.Stderr, "init: destination required\n")
			os.Exit(1)
		}

		log.Println("creating repository", *destination)
		if err := createRepository(mustClient(), *destination); err != nil {
			fmt.Fprintf(os.Stderr, "create repository: error for %s %v", *destination, err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}

func createRepository(client *esclient.Client, destination string) error {
	bucket := fmt.Sprintf("%s%s", viper.GetString("bucket"), destination)
	return client.CreateSnapshotRepository(destination, esclient.RepositorySettings{
		Bucket:               bucket,
		Region:               viper.GetString("AZ"),
		ServerSideEncryption: viper.GetBool("server_side_encryption"),
		Protocol:             viper.GetString("protocol"),
	})
}
