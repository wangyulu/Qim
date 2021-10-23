package container

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"errors"

	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/naming"
	"jinv/kim/tcp"
	"jinv/kim/wire"
	"jinv/kim/wire/pkt"
)

const (
	stateUninitialized = iota
	stateInitialized
	stateStarted
	stateClosed
)

const (
	StateYoung = "young"
	StateAdult = "adult"
)

const (
	KeyServiceState = "service_state"
)

type Container struct {
	sync.RWMutex
	Naming     naming.Naming
	Srv        kim.Server
	state      uint32
	srvClients map[string]ClientMap
	selector   Selector
	dialer     kim.Dialer
	deps       map[string]struct{}
}

var log = logger.WithField("module", "container")

var c = &Container{
	state:    0,
	selector: &HashSelector{},
	deps:     make(map[string]struct{}),
}

func Default() *Container {
	return c
}

func Init(srv kim.Server, deps ...string) error {
	if !atomic.CompareAndSwapUint32(&c.state, stateUninitialized, stateInitialized) {
		return errors.New("has Initialized")
	}

	c.Srv = srv

	for _, dep := range deps {
		if _, ok := c.deps[dep]; ok {
			continue
		}

		c.deps[dep] = struct{}{}
	}

	log.WithField("func", "Init").Infof("srv %s:%s - deps %v", srv.ServiceID(), srv.ServiceName(), c.deps)

	c.srvClients = make(map[string]ClientMap, len(deps))

	return nil
}

func SetDialer(dialer kim.Dialer) {
	c.dialer = dialer
}

func SetSelector(selector Selector) {
	c.selector = selector
}

func SetServiceNaming(nm naming.Naming) {
	c.Naming = nm
}

func Start() error {
	if c.Naming == nil {
		return fmt.Errorf("naming is nil")
	}

	if !atomic.CompareAndSwapUint32(&c.state, stateInitialized, stateStarted) {
		return errors.New("has started")
	}

	// 1. 启动Server
	go func(srv kim.Server) {
		if err := srv.Start(); err != nil {
			log.Errorln(err)
		}
	}(c.Srv)

	// 2. 与依赖的服务建立连接
	for service := range c.deps {
		go func(service string) {
			err := connectToService(service)
			if err != nil {
				log.Errorln(err)
			}
		}(service)
	}

	// 3. 服务注册
	if c.Srv.PublicAddress() != "" && c.Srv.PublicPort() != 0 {
		err := c.Naming.Register(c.Srv)
		if err != nil {
			log.Errorln(err)
		}
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	log.Infoln("shutdown", <-c)

	// 4. 服务退出
	return shutdown()
}

func shutdown() error {
	if !atomic.CompareAndSwapUint32(&c.state, stateStarted, stateClosed) {
		return errors.New("has closed")
	}

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()

	// 1. 优雅关闭服务器
	if err := c.Srv.Shutdown(ctx); err != nil {
		log.Error(err)
	}

	// 2. 从注册中心注销服务
	if err := c.Naming.Deregister(c.Srv.ServiceID()); err != nil {
		log.Warn(err)
	}

	// 3. 退订服务变更
	for dep := range c.deps {
		_ = c.Naming.Unsubscribe(dep)
	}

	log.Infoln("shutdown")

	return nil
}

// 在逻辑服务中，消息是被上层业务手动调用容器中的Push方法，把消息发送给网关的
// todo
func Push(server string, p *pkt.LogicPkt) error {
	p.AddStringMeta(wire.MetaDestServer, server)

	return c.Srv.Push(server, pkt.Marshal(p))
}

func Forward(serviceName string, packet *pkt.LogicPkt) error {
	if packet == nil {
		return errors.New("packet is nil")
	}

	if packet.Command == "" {
		return errors.New("command is empty in packet")
	}

	if packet.ChannelId == "" {
		return errors.New("ChannelId is empty in packet")
	}

	return ForwardWithSelector(serviceName, packet, c.selector)
}

func ForwardWithSelector(serviceName string, packet *pkt.LogicPkt, selector Selector) error {
	cli, err := lookup(serviceName, &packet.Header, selector)
	if err != nil {
		return err
	}

	packet.AddStringMeta(wire.MetaDestServer, c.Srv.ServiceID())

	log.Debugf("forward message to %v with %s", cli.ServiceID(), &packet.Header)

	return cli.Send(pkt.Marshal(packet))
}

func lookup(serviceName string, header *pkt.Header, selector Selector) (kim.Client, error) {
	clients, ok := c.srvClients[serviceName]
	if !ok {
		return nil, fmt.Errorf("service %s not found", serviceName)
	}

	// 只获取状态为StateAdult的服务
	srvs := clients.Services(KeyServiceState, StateAdult)
	if len(srvs) == 0 {
		return nil, fmt.Errorf("no services found for %s", serviceName)
	}

	id := selector.Lookup(header, srvs)
	if cli, ok := clients.Get(id); ok {
		return cli, nil
	}

	return nil, fmt.Errorf("no client found")
}

func connectToService(serviceName string) error {
	clients := NewClients(10)

	c.srvClients[serviceName] = clients

	// 1. 首先 watch 服务的新增
	delay := time.Second * 10

	err := c.Naming.Subscribe(serviceName, func(services []kim.ServiceRegistration) {
		for _, service := range services {
			if _, ok := clients.Get(service.ServiceID()); ok {
				continue
			}

			log.WithField("func", "connectToService").Infof("watch a new service: %v", service)

			service.GetMeta()[KeyServiceState] = StateYoung

			go func() {
				time.Sleep(delay) // todo ？

				service.GetMeta()[KeyServiceState] = StateAdult
			}()

			_, err := buildClient(clients, service)
			if err != nil {
				logger.Warn(err)
			}
		}
	})

	if err != nil {
		return err
	}

	// 2. 再查询已经存在的服务
	services, err := c.Naming.Find(serviceName)
	if err != nil {
		return err
	}

	log.Info("find service ", services)

	for _, service := range services {
		// 标记服务的状态为 StateAudit todo 这里为何放在这里处理服务的状态呢
		service.GetMeta()[KeyServiceState] = StateAdult

		_, err := buildClient(clients, service)
		if err != nil {
			logger.Warn(err)
		}
	}

	return nil
}

func buildClient(clients ClientMap, service kim.ServiceRegistration) (kim.Client, error) {
	c.Lock()
	defer c.Unlock()

	var (
		id   = service.ServiceID()
		name = service.ServiceName()
		meta = service.GetMeta()
	)

	// 1. 检测连接是否已经存在
	if _, ok := clients.Get(id); ok {
		return nil, nil
	}

	// 2. 服务之间只允许使用 tcp 协议
	if service.GetProtocol() != string(wire.ProtocolTCP) {
		return nil, fmt.Errorf("unexpected service protocol: %s", service.GetProtocol())
	}

	// 3. 构造客户端并建立连接
	cli := tcp.NewClientWithProps(id, name, meta, tcp.ClientOptions{
		Heartbeat: kim.DefaultHeartbeat,
		ReadWait:  kim.DefaultReadWait,
		WriteWait: kim.DefaultWriteWait,
	})

	if c.dialer == nil {
		return nil, fmt.Errorf("dialer is nil")
	}

	cli.SetDialer(c.dialer)

	err := cli.Connect(service.DialURL())
	if err != nil {
		return nil, err
	}

	// 4. 读取消息
	go func(cli kim.Client) {
		err := readLoop(cli)
		if err != nil {
			log.Debug(err)
		}
		clients.Remove(id)

		cli.Close()
	}(cli)

	// 5. 添加到客户端集合中去
	clients.Add(cli)

	return cli, nil
}

func readLoop(cli kim.Client) error {
	log := logger.WithFields(logger.Fields{"module": "container", "func": "readLoop"})

	log.Infof("readLoop started of %s %s", cli.ServiceID(), cli.ServiceName())

	for {
		frame, err := cli.Read()
		if err != nil {
			return err
		}

		if frame.GetOpCode() != kim.OpBinary {
			continue
		}

		buf := bytes.NewBuffer(frame.GetPayload())

		packet, err := pkt.MustReadLogicPkt(buf)
		if err != nil {
			log.Info(err) // todo 这里直接打印一条info信息是否合适呢
			continue
		}

		err = pushMessage(packet)
		if err != nil {
			log.Info(err)
		}
	}
}

// todo
func pushMessage(packet *pkt.LogicPkt) error {
	server, _ := packet.GetMeta(wire.MetaDestServer)
	if server != c.Srv.ServiceID() {
		return fmt.Errorf("dest_server is not incorrect, %s != %s", server, c.Srv.ServiceID())
	}

	channels, ok := packet.GetMeta(wire.MetaDestChannels)
	if !ok {
		return fmt.Errorf("dest_channnels is nil")
	}

	channelIds := strings.Split(channels.(string), ",")

	packet.DelMeta(wire.MetaDestServer) // todo 这里为什么要删除呢
	packet.DelMeta(wire.MetaDestChannels)

	payload := pkt.Marshal(packet)

	log.Debugf("Push to %v %v", channelIds, packet)

	for _, channel := range channelIds {
		err := c.Srv.Push(channel, payload)
		if err != nil {
			log.Debug(err)
		}
	}

	return nil
}
