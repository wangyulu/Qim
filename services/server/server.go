package server

import (
	"context"

	"github.com/spf13/cobra"
	"jinv/kim"
	"jinv/kim/container"
	"jinv/kim/logger"
	"jinv/kim/naming"
	"jinv/kim/naming/consul"
	"jinv/kim/services/server/conf"
	"jinv/kim/services/server/handler"
	"jinv/kim/services/server/serv"
	"jinv/kim/storage"
	"jinv/kim/tcp"
	"jinv/kim/wire"
)

type ServerStartOptions struct {
	config      string
	serviceName string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Start a server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./server/conf.yaml", "config file")
	cmd.PersistentFlags().StringVarP(&opts.serviceName, "serviceName", "s", "chat", "defined a service name, option is login or chat")

	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	// 1. 配置初始化
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}

	_ = logger.Init(logger.Settings{Level: "trace"})

	// 2. 路由
	r := kim.NewRouter()

	loginHandler := handler.NewLoginHandler()
	r.Handle(wire.CommandLoginSignIn, loginHandler.DoSysLogin)
	r.Handle(wire.CommandLoginSignOut, loginHandler.DoSysLogout)

	// 3. 会话存储
	rdb, err := conf.InitRedis(config.RedisAddrs, "")
	if err != nil {
		return err
	}

	cache := storage.NewRedisStorage(rdb)

	servHandler := serv.NewServHandler(r, cache)

	service := &naming.DefaultService{
		Id:       config.ServiceID,
		Name:     opts.serviceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: string(wire.ProtocolTCP),
		Tags:     config.Tags,
	}

	srv := tcp.NewServer(config.Listen, service)

	srv.SetReadWait(kim.DefaultReadWait)
	srv.SetAcceptor(servHandler)
	srv.SetMessageListener(servHandler)
	srv.SetStateListener(servHandler)

	if err := container.Init(srv); err != nil {
		return err
	}

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}

	container.SetServiceNaming(ns)

	return container.Start()
}
