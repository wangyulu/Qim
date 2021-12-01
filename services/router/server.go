package router

import (
	"context"
	"path"

	"github.com/kataras/iris/v12"
	"github.com/spf13/cobra"
	"jinv/kim/logger"
	"jinv/kim/naming/consul"
	"jinv/kim/services/router/apis"
	"jinv/kim/services/router/conf"
	"jinv/kim/services/router/ipregion"
)

type ServerStartOptions struct {
	config string
	data   string
}

func NewStartServerCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "router",
		Short: "start a router",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./router/conf.yaml", "config file")
	cmd.PersistentFlags().StringVarP(&opts.data, "data", "d", "./router/data", "data path")

	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}

	_ = logger.Init(logger.Settings{
		Level:    config.LogLevel,
		Filename: "./data/router.log",
	})

	mappings, err := conf.LoadMapping(path.Join(opts.data, "mapping.json"))
	if err != nil {
		return err
	}
	logger.Infof("load mappings - %v", mappings)

	regions, err := conf.LoadRegions(path.Join(opts.data, "regions.json"))
	if err != nil {
		return err
	}
	logger.Infof("load regions - %v", regions)

	ipRegion, err := ipregion.NewIp2region(path.Join(opts.data, "ip2region.db"))
	if err != nil {
		return err
	}

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}

	router := apis.RouterApi{
		Naming:   ns,
		IpRegion: ipRegion,
		Config: conf.Router{
			Mapping: mappings,
			Regions: regions,
		},
	}

	app := iris.Default()

	// 这里其实也不需要，因为没有注册到 Consul
	app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.WriteString("ok")
	})

	routerAPI := app.Party("/api/lookup")
	{
		routerAPI.Get("/:token", router.Lookup)
	}

	return app.Listen(config.Listen, iris.WithOptimizations)
}
