package service

import (
	"context"
	"fmt"
	"hash/crc32"

	"github.com/kataras/iris/v12"
	"github.com/spf13/cobra"
	"jinv/kim/logger"
	"jinv/kim/naming"
	"jinv/kim/naming/consul"
	"jinv/kim/services/service/conf"
	"jinv/kim/services/service/database"
	"jinv/kim/services/service/handler"
	"jinv/kim/wire"
)

type ServerStartOptions struct {
	config string
}

func NewServerStartCMD(ctx context.Context, version string) *cobra.Command {
	opts := &ServerStartOptions{}

	cmd := &cobra.Command{
		Use:   "royal",
		Short: "Start a rpc service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunStartStart(ctx, opts, version)
		},
	}

	cmd.PersistentFlags().StringVarP(&opts.config, "config", "c", "./service/conf.yaml", "Config File")

	return cmd
}

func RunStartStart(ctx context.Context, opts *ServerStartOptions, version string) error {
	config, err := conf.Init(opts.config)
	if err != nil {
		return err
	}

	_ = logger.Init(logger.Settings{
		Level:    config.LogLevel,
		Filename: "./data/royal.log",
	})

	baseDb, err := database.InitMysqlDb(config.BaseDb)
	if err != nil {
		return err
	}
	_ = baseDb.AutoMigrate(&database.Group{}, &database.GroupMember{})

	messageDb, err := database.InitMysqlDb(config.MessageDb)
	if err != nil {
		return err
	}
	_ = messageDb.AutoMigrate(&database.MessageIndex{}, &database.MessageContent{})

	if config.NodeID == 0 {
		config.NodeID = int64(HashCode(config.ServiceID))
	}

	idGen, err := database.NewIDGenerator(config.NodeID)
	if err != nil {
		return err
	}

	redisDb, err := conf.InitRedis(config.RedisAddrs, "")
	if err != nil {
		return err
	}

	ns, err := consul.NewNaming(config.ConsulURL)
	if err != nil {
		return err
	}

	_ = ns.Register(&naming.DefaultService{
		Id:       config.ServiceID,
		Name:     wire.SNService,
		Address:  config.PublicAddress,
		Port:     config.PublicPort,
		Protocol: "http",
		Tags:     config.Tags,
		Meta: map[string]string{
			consul.KeyHealthURL: fmt.Sprintf("http://%s:%d/health", config.PublicAddress, config.PublicPort),
		},
	})
	defer func() {
		_ = ns.Deregister(config.ServiceID)
	}()

	serviceHandler := &handler.ServiceHandler{
		BaseDb:    baseDb,
		MessageDb: messageDb,
		Cache:     redisDb,
		IdGen:     idGen,
	}

	// req  & resp log
	ac := conf.MakeAccessLog()
	defer ac.Close()

	app := newApp(serviceHandler)
	app.UseRouter(ac.Handler)
	app.UseRouter(setAllowedResponses)

	return app.Listen(config.Listen, iris.WithOptimizations)
}

func newApp(serviceHandler *handler.ServiceHandler) *iris.Application {
	app := iris.Default()

	app.Get("/health", func(ctx iris.Context) {
		_, _ = ctx.WriteString("ok")
	})

	messageAPI := app.Party("/api/:app/message")
	{
		messageAPI.Post("/user", serviceHandler.InsertUserMessage)
		messageAPI.Post("/group", serviceHandler.InsertGroupMessage)
		messageAPI.Post("/ack", serviceHandler.MessageAck)
	}

	groupAPI := app.Party("/api/:app/group")
	{
		groupAPI.Get("/:id", serviceHandler.GroupGet)
		groupAPI.Post("", serviceHandler.GroupCreate)
		groupAPI.Post("/member", serviceHandler.GroupJoin)
		groupAPI.Delete("/member", serviceHandler.GroupQuit)
		groupAPI.Get("/members/:id", serviceHandler.GroupMembers)
	}

	offlineAPI := app.Party("/api/:app/offline")
	{
		offlineAPI.Use(iris.Compression)
		offlineAPI.Post("/index", serviceHandler.GetOfflineMessageIndex)
		offlineAPI.Post("/content", serviceHandler.GetOfflineMessageContent)
	}

	return app
}

func setAllowedResponses(ctx iris.Context) {
	// Indicate that the Server can send JSON, XML, YAML and MessagePack for this request.
	ctx.Negotiation().JSON().Protobuf().MsgPack()
	// Add more, allowed by the server format of responses, mime types here...

	// If client is missing an "Accept: " header then default it to JSON.
	ctx.Negotiation().Accept.JSON()

	ctx.Next()
}

func HashCode(key string) uint32 {
	hash := crc32.NewIEEE()

	_, _ = hash.Write([]byte(key))

	return hash.Sum32() % 1000
}
