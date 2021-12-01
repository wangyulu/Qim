package gateway

import (
	"context"
	"time"

	"github.com/spf13/cobra"
	"jinv/kim"
	"jinv/kim/container"
	"jinv/kim/logger"
	"jinv/kim/naming"
	"jinv/kim/naming/consul"
	"jinv/kim/services/gateway/conf"
	"jinv/kim/services/gateway/serv"
	"jinv/kim/websocket"
	"jinv/kim/wire"
)

type ServerStartOptions struct {
	config   string
	protocol string
}

func NewServerStartCmd(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Start a gateway",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServerStart(ctx, opts, version)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./gateway/conf.yaml", "config file")
	cmd.PersistentFlags().StringVarP(&opts.protocol, "protocol", "p", "ws", "protocol of ws or tcp")

	return cmd
}

func RunServerStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}

	_ = logger.Init(logger.Settings{
		Level:    config.LogLevel,
		Filename: "./data/gateway.log",
	})

	meta := make(map[string]string)
	meta["domain"] = config.Domain

	handler := &serv.Handler{
		ServiceID: config.ServiceID,
	}

	var srv kim.Server

	service := &naming.DefaultService{
		Id:       config.ServiceID,
		Name:     config.ServiceName,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: opts.protocol,
		Tags:     config.Tags,
		Meta:     meta,
	}

	if opts.protocol == "ws" {
		srv = websocket.NewServer(config.Listen, service)
	}

	srv.SetReadWait(time.Minute * 2)
	srv.SetAcceptor(handler)
	srv.SetMessageListener(handler)
	srv.SetStateListener(handler)

	if err := container.Init(srv, wire.SNChat, wire.SNLogin); err != nil {
		return err
	}

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}

	container.SetServiceNaming(ns)

	container.SetDialer(serv.NewDialer(config.ServiceID))

	return container.Start()
}
