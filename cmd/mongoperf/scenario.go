package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"text/template"
	"time"

	"mongo-tester/internal/mongodb"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	reportTemplate = `
===============================
  Scenario Report
===============================
  This report was produced
  using mongoperf {{ .Version }}
-------------------------------
  Config
-------------------------------
  URI:       {{ .URI }}
  Database   {{ .Database }}
  Collection {{ .Collection }}
  Parallel   {{ .Parallel }}
-------------------------------
  Queries
-------------------------------
{{ with .Queries -}}
{{ range $name, $elem := . -}}
{{ block "query" . }}{{ end }}
{{- end }}
{{- end }}
===============================
`

	queryBlock = `
{{ define "query" }}
  > Name:        {{ .Name }}
    Action:      {{ .Action }}
    {{ with .Result -}}
    Successful:  {{ .Successful }}
    TotalTime:   {{ .TotalTime }}
    TotalChange: {{ .TotalChange }}
    Error:       {{ if .Error }}{{ .Error.Error }}{{ else }}nil{{ end }}
    {{- end }}
{{ end }}
`
)

// Report .
type Report struct {
	Version    string
	URI        string
	Database   string
	Collection string
	Parallel   int
	Queries    map[string]*Query
}

// Query .
type Query struct {
	Name   string
	Action string
	Result struct {
		Successful  bool
		TotalTime   time.Duration
		TotalChange int
		Error       error
	}
}

func queriesFromScenario(sr *mongodb.ScenarioReport) map[string]*Query {
	queries := make(map[string]*Query)
	for name, result := range sr.QueryResult {
		query := &Query{
			Name:   name,
			Action: *result.Query.Action,
			Result: struct {
				Successful  bool
				TotalTime   time.Duration
				TotalChange int
				Error       error
			}{
				Successful:  result.Error == nil,
				TotalTime:   result.End.Sub(result.Start),
				TotalChange: result.TotalChange,
				Error:       result.Error,
			},
		}
		queries[name] = query
	}
	return queries
}

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

			if scenario.Scenario.Parallel < 1 {
				return fmt.Errorf("Scenario.Parallel must be greater than 0")
			}

			if uri == "" {
				uri = "mongodb://localhost:27017"
			}
			logger.Printf("connecting to: %v", uri)

			client := mongodb.NewClient(uri, mongodb.WithLogger(logger))

			interrupt := make(chan os.Signal, 1)
			signal.Notify(interrupt, os.Interrupt)

			ctx, cancelFn := context.WithCancel(context.Background())

			done := make(chan struct{}, 0)
			scenarioReport := mongodb.NewScenarioReport()

			go client.RunScenario(ctx, scenario.Scenario, scenarioReport, done)

			select {
			case <-interrupt:
			case <-done:
			}
			cancelFn()

			report := &Report{
				Version:    cmd.Parent().Version,
				URI:        uri,
				Database:   *scenario.Scenario.Database,
				Collection: *scenario.Scenario.Collection,
				Parallel:   scenario.Scenario.Parallel,
				Queries:    queriesFromScenario(scenarioReport),
			}

			t, err := parseTemplates("report-template", reportTemplate, queryBlock)
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
	return cmd
}

func parseTemplates(name string, tmpl ...string) (*template.Template, error) {
	if len(tmpl) == 0 {
		return nil, fmt.Errorf("no templates provided")
	}
	var t *template.Template
	for idx, tStr := range tmpl {
		if t == nil {
			t = template.New(fmt.Sprintf("%v-%d", name, idx))
		}
		var err error
		t, err = t.Parse(tStr)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
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
