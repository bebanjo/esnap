package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	esclient "github.com/bebanjo/esnap/internal/es"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "esnap",
	Short: "Manage Elasticsearch snapshots and take a nap",
	Long:  ``,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

var destination *string
var createRepositoryTake *bool
var originRestore, snapshot *string
var allIndices, fresh, aliased *bool
var age *int

var (
	esClient     *esclient.Client
	esClientErr  error
	esClientOnce sync.Once
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.esnap.yaml)")
	destination = RootCmd.PersistentFlags().StringP("destination", "d", "", "Destination for the command action")
}

func initConfig() {
	log.SetOutput(os.Stdout)

	viper.SetDefault("bucket", "my-bucket")
	viper.SetDefault("AZ", "eu-west-1")
	viper.SetDefault("protocol", "https")
	viper.SetDefault("server_side_encryption", true)
	viper.SetDefault("elasticsearch_url", "http://localhost:9200")
	viper.SetDefault("elasticsearch_username", "")
	viper.SetDefault("elasticsearch_password", "")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName(".esnap")
		viper.AddConfigPath("$HOME")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	_ = viper.BindEnv("elasticsearch_url", "ES_URL")
	_ = viper.BindEnv("elasticsearch_username", "ES_USERNAME")
	_ = viper.BindEnv("elasticsearch_password", "ES_PASSWORD")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func configuredClient() (*esclient.Client, error) {
	esClientOnce.Do(func() {
		addresses := splitAndTrim(viper.GetString("elasticsearch_url"))
		esClient, esClientErr = esclient.NewClient(esclient.Config{
			Addresses: addresses,
			Username:  viper.GetString("elasticsearch_username"),
			Password:  viper.GetString("elasticsearch_password"),
		})
	})
	return esClient, esClientErr
}

func mustClient() *esclient.Client {
	client, err := configuredClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "elasticsearch client: %v\n", err)
		os.Exit(1)
	}
	return client
}

func splitAndTrim(value string) []string {
	parts := strings.Split(value, ",")
	addresses := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			addresses = append(addresses, part)
		}
	}
	return addresses
}
