package main

import (
	"context"
	"flag"

	"github.com/spf13/cobra"
	"jinv/kim/logger"
	"jinv/kim/services/gateway"
	"jinv/kim/services/router"
	"jinv/kim/services/server"
	"jinv/kim/services/service"
)

const version = "v1"

func main() {
	flag.Parse()

	root := &cobra.Command{
		Use:     "kim",
		Version: version,
		Short:   "King IM Cloud",
	}

	ctx := context.Background()

	root.AddCommand(gateway.NewServerStartCmd(ctx, version))
	root.AddCommand(server.NewServerStartCmd(ctx, version))
	root.AddCommand(service.NewServerStartCMD(ctx, version))
	root.AddCommand(router.NewStartServerCmd(ctx, version))

	if err := root.Execute(); err != nil {
		logger.WithError(err).Fatal("Could not run command")
	}
}
