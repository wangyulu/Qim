package server

import (
	"context"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"jinv/kim"
	"jinv/kim/container"
	"jinv/kim/logger"
	"jinv/kim/middleware"
	"jinv/kim/naming"
	"jinv/kim/naming/consul"
	"jinv/kim/services/server/conf"
	"jinv/kim/services/server/handler"
	"jinv/kim/services/server/serv"
	"jinv/kim/services/server/service"
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

	var groupService service.Group
	var messageService service.Message

	if strings.TrimSpace(config.RoyalURL) != "" {
		groupService = service.NewGroupService(config.RoyalURL)
		messageService = service.NewMessageService(config.RoyalURL)
	} else {
		// todo 这是个啥啊
		srvRecord := &resty.SRVRecord{
			Domain:  "consul",
			Service: wire.SNService,
		}

		groupService = service.NewGroupServiceWithSRV("http", srvRecord)
		messageService = service.NewMessageServiceWithSRV("http", srvRecord)
	}

	// 2. 路由
	r := kim.NewRouter()
	r.Use(middleware.Recover())

	// login
	loginHandler := handler.NewLoginHandler()
	r.Handle(wire.CommandLoginSignIn, loginHandler.DoSysLogin)
	r.Handle(wire.CommandLoginSignOut, loginHandler.DoSysLogout)

	// talk
	chatHandler := handler.NewChatHandler(messageService, groupService)
	r.Handle(wire.CommandChatUserTalk, chatHandler.DoUserTalk)
	r.Handle(wire.CommandChatGroupTalk, chatHandler.DoGroupTalk)
	r.Handle(wire.CommandChatTalkAck, chatHandler.DoTalkAck)

	// group
	groupHandler := handler.NewGroupHandler(groupService)
	r.Handle(wire.CommandGroupCreate, groupHandler.DoCreate)
	r.Handle(wire.CommandGroupJoin, groupHandler.DoJoin)
	r.Handle(wire.CommandGroupQuit, groupHandler.DoQuit)
	r.Handle(wire.CommandGroupDetail, groupHandler.DoDetail)

	// offline
	offlineHandler := handler.NewOfflineHandler(messageService)
	r.Handle(wire.CommandOfflineIndex, offlineHandler.DoSyncIndex)
	r.Handle(wire.CommandOfflineContext, offlineHandler.DoSyncContent)

	// 3. 会话存储
	rdb, err := conf.InitRedisClusterV2(config.RedisClusterAddrs, "")
	if err != nil {
		return err
	}

	cache := storage.NewRedisClusterStorageV2(rdb)

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
