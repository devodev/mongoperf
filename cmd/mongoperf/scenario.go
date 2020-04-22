package main

import (
	"context"
	"mongoperf/internal/scenario"
	"mongoperf/internal/scenario/query"
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
		Short: "Run a scenario.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// PARSE CONFIG
			cfgFile := args[0]
			config, err := scenario.ParseConfigFile(cfgFile)
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
			client, err := scenario.NewClient(context.TODO(), uri, scenario.WithLogger(logger))
			if err != nil {
				return err
			}

			// SETUP CHANNEL, CONTEXT AND HANDLERS
			interruptCh := make(chan os.Signal, 1)
			signal.Notify(interruptCh, os.Interrupt)

			doneCh := make(chan struct{}, 0)
			resultCh := make(chan *query.Result, 0)

			// START INTERRUPT HANDLER
			ctx, cancelCtx := context.WithCancel(context.Background())
			go func() {
				defer cancelCtx()
				select {
				case <-interruptCh:
				case <-doneCh:
				}
			}()

			// START RESULT PROCESSING
			queries := make(map[string]*scenario.ReportQuery)
			go func() {
				for result := range resultCh {
					rq, ok := queries[*result.Query.Name]
					if !ok {
						rq = scenario.NewReportQuery(*result.Query.Name, string(*result.Query.Action))
					}
					delta := result.End.Sub(result.Start)
					rq.Update(delta, result.TotalChange, result.Error)
					queries[*result.Query.Name] = rq
				}
				close(doneCh)
			}()

			// START SCENARIO
			go client.RunScenario(ctx, config, resultCh)

			// WAIT ON COMPLETION
			select {
			case <-doneCh:
			}

			// GENERATE REPORT
			report := &scenario.Report{
				Version:    cmd.Parent().Version,
				URI:        uri,
				Database:   *config.Database,
				Collection: *config.Collection,
				Parallel:   *config.Parallel,
				Queries:    queries,
			}

			templates := []string{scenario.ReportTemplate, scenario.ConfigBlock, scenario.QueryBlock}
			t, err := scenario.ParseTemplates("report-template", templates...)
			if err != nil {
				return err
			}
			if err := t.Execute(defaultOutput, report); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&uri, "uri", "", "MongoDB URI connection string.")
	cmd.Flags().BoolVar(&isDebug, "debug", false, "Set logger level to DEBUG.")
	return cmd
}
