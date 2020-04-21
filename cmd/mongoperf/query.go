package main

import (
	"context"
	"os"

	"mongo-tester/internal/mongodb"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCommandQuery() *cobra.Command {
	var (
		cfgFile    string
		uri        string
		db         string
		collection string
	)
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Query mongodb and report statistics.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logrus.New()

			config, err := getConfig(cfgFile)
			if err != nil {
				return err
			}
			logger.Printf("using config: %+v", config)

			if uri == "" {
				uri = "mongodb://localhost:27017"
			}
			if db == "" {
				db = "testdb"
			}
			if collection == "" {
				collection = "testcol"
			}
			logger.Printf("connecting to: %v", uri)
			logger.Printf("using database: %v", db)
			logger.Printf("using collection: %v", collection)

			client := mongodb.New(uri, mongodb.WithLogger(logger))

			if err := client.RunDemo(context.TODO(), db, collection); err != nil {
				logger.Fatal(err)
			}

			return nil
		},
	}
	cmd.Flags().StringVar(&cfgFile, "config", "", "Configuration file containing scenario declarations. Default is: $CWD/.mongo-test.yaml")
	cmd.Flags().StringVar(&uri, "uri", "", "MongoDB URI connection string.")
	cmd.Flags().StringVar(&db, "db", "", "MongoDB database.")
	cmd.Flags().StringVar(&collection, "collection", "", "MongoDB collection.")
	return cmd
}

func getConfig(cfgFile string) (*config, error) {
	viperInstance := viper.New()
	if cfgFile != "" {
		viperInstance.SetConfigFile(cfgFile)
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		viperInstance.AddConfigPath(wd)
		viperInstance.SetConfigName(".mongo-test")
		viperInstance.SetConfigType("yaml")
	}

	viperInstance.AutomaticEnv()

	err := viperInstance.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config config
	if err := viperInstance.UnmarshalExact(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

type config struct {
}
