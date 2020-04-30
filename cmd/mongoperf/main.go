package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	loggerOutput  = os.Stderr
	defaultOutput = os.Stdout
)

func writeOut(line string) {
	fmt.Fprintln(defaultOutput, line)
}

func newCommandRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "mongoperf",
		Short:   "Run performance tests scenarios on a mongodb instance or cluster.",
		Version: "0.1.1",
	}
	cmd.AddCommand(
		newCommandScenario(),
	)
	return cmd
}

// Execute executes the root command.
func Execute() error {
	rootCmd := newCommandRoot()
	return rootCmd.Execute()
}

func main() {
	if err := Execute(); err != nil {
		writeOut(err.Error())
		os.Exit(1)
	}
}
