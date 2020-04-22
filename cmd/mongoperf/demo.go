package main

import (
	"context"

	"mongo-tester/internal/mongodb"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newCommandDemo() *cobra.Command {
	var (
		uri        string
		db         string
		collection string
	)
	cmd := &cobra.Command{
		Use:   "demo",
		Short: "Run small demo that inserts, update and delete entries.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logrus.New()

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

			client, err := mongodb.NewClient(context.TODO(), uri, mongodb.WithLogger(logger))
			if err != nil {
				return err
			}
			if err := client.RunDemo(context.TODO(), db, collection); err != nil {
				logger.Fatal(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&uri, "uri", "", "MongoDB URI connection string.")
	cmd.Flags().StringVar(&db, "db", "", "MongoDB database.")
	cmd.Flags().StringVar(&collection, "collection", "", "MongoDB collection.")
	return cmd
}

type config struct {
	Scenarios []*mongodb.ScenarioConfig
}
