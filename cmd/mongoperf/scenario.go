package main

import (
	"context"

	"mongo-tester/internal/mongodb"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newCommandScenario() *cobra.Command {
	var (
		uri string
	)
	cmd := &cobra.Command{
		Use:   "scenario [scenario-file]",
		Short: "Run a scenario.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			logger := logrus.New()

			cfgFile := args[0]
			scenario, err := getScenarioConfig(cfgFile)
			if err != nil {
				return err
			}

			if uri == "" {
				uri = "mongodb://localhost:27017"
			}
			logger.Printf("connecting to: %v", uri)

			client := mongodb.New(uri, mongodb.WithLogger(logger))

			if err := client.RunScenario(context.TODO(), scenario.Scenario); err != nil {
				logger.Fatal(err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&uri, "uri", "", "MongoDB URI connection string.")
	return cmd
}

func getScenarioConfig(cfgFile string) (*scenarioConfig, error) {
	viperInstance := viper.New()
	viperInstance.SetConfigFile(cfgFile)

	err := viperInstance.ReadInConfig()
	if err != nil {
		return nil, err
	}

	var config scenarioConfig
	if err := viperInstance.UnmarshalExact(&config); err != nil {
		return nil, err
	}
	return &config, nil
}

type scenarioConfig struct {
	Scenario *mongodb.Scenario
}
