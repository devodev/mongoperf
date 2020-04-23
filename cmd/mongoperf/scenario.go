package main

import (
	"context"
	"mongoperf/internal/client"
	"os"
	"os/signal"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newCommandScenario() *cobra.Command {
	var (
		uri     string
		isDebug bool
	)
	cmd := &cobra.Command{
		Use:   "scenario [scenario-file]",
		Short: "Run a client.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// PARSE CONFIG
			cfgFile := args[0]
			scenario, err := client.ParseScenarioFile(cfgFile)
			if err != nil {
				return err
			}

			// VALIDATE COMMAND LINE ARGS
			if uri == "" {
				uri = "mongodb://localhost:27017"
			}

			// CREATE LOGGER
			logger := logrus.New()
			if isDebug {
				logger.SetLevel(logrus.DebugLevel)
			}

			// START CLIENT
			logger.Printf("connecting to: %v", uri)
			c, err := client.New(context.TODO(), uri, client.WithLogger(logger))
			if err != nil {
				return err
			}

			// SETUP INTERRUPT HANDLER
			interruptCh := getInterruptCh()
			ctx, cancelCtx := context.WithCancel(context.Background())

			go func() {
				select {
				case <-interruptCh:
					cancelCtx()
				}
			}()

			// RUN SCENARIO
			queryResults, err := c.RunScenario(ctx, scenario)

			// GENERATE REPORT
			report := client.NewReport(cmd.Parent().Version, uri, scenario, queryResults)
			if err := client.GenerateReport(defaultOutput, report); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&uri, "uri", "", "MongoDB URI connection string.")
	cmd.Flags().BoolVar(&isDebug, "debug", false, "Set logger level to DEBUG.")
	return cmd
}

func getInterruptCh() <-chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	return sigChan
}
