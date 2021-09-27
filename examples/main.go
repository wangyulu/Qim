package main

import (
	"flag"

	"github.com/spf13/cobra"
	"jinv/kim/examples/mock"
	"jinv/kim/logger"
)

func main() {
	flag.Parse()

	root := cobra.Command{
		Use:   "kim",
		Short: "im",
	}

	// mock
	root.AddCommand(mock.NewServerCmd())
	root.AddCommand(mock.NewClientCmd())

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
