package main

import (
	"flag"

	"context"

	"github.com/spf13/cobra"
	"jinv/kim/examples/kimbench"
	"jinv/kim/examples/mock"
	"jinv/kim/logger"
)

func main() {
	flag.Parse()

	root := cobra.Command{
		Use:   "kim",
		Short: "tools",
	}

	ctx := context.Background()

	// mock
	root.AddCommand(mock.NewServerCmd(ctx))
	root.AddCommand(mock.NewClientCmd(ctx))
	root.AddCommand(kimbench.NewBenchmarkCmd(ctx))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
