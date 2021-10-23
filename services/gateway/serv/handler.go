package serv

import (
	"bytes"
	"fmt"
	"regexp"
	"time"

	"jinv/kim"
	"jinv/kim/container"
	"jinv/kim/logger"
	"jinv/kim/wire"
	"jinv/kim/wire/pkt"
	"jinv/kim/wire/token"
)

var log = logger.WithFields(logger.Fields{"service": "gateway", "pkg": "serv"})

type Handler struct {
	ServiceID string
}

func (h *Handler) Accept(conn kim.Conn, timeout time.Duration) (string, error) {
	log := logger.WithFields(logger.Fields{
		"ServiceID": h.ServiceID,
		"module":    "Handler",
		"handler":   "Accpet",
	})

	log.Infoln("enter")

	// 1. 读取登录包
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return "", err
	}

	frame, err := conn.ReadFrame()
	if err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(frame.GetPayload())
	req, err := pkt.MustReadLogicPkt(buf)
	if err != nil {
		return "", nil
	}

	// 2. 必须是登录包
	if req.Command != wire.CommandLoginSignIn {
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_InvalidCommand

		_ = conn.WriteFrame(kim.OpBinary, pkt.Marshal(resp))

		return "", fmt.Errorf("must be a InvalidCommand command")
	}

	// 3. 反序列化Body
	var login pkt.LoginReq
	if err = req.ReadBody(&login); err != nil {
		return "", err
	}

	// 4. 使用默认的DefaultSecret解析Token
	tk, err := token.Parse(token.DefaultSecret, login.Token)
	if err != nil {
		// 5. 如果Token无效，就返回SDK一个Unauthorized消息
		resp := pkt.NewFrom(&req.Header)
		resp.Status = pkt.Status_Unauthorized

		_ = conn.WriteFrame(kim.OpBinary, pkt.Marshal(resp))

		return "", err
	}

	// 6. 生成一个全局唯一的ChannelID
	id := generateChannelID(h.ServiceID, tk.Account)

	req.ChannelId = id
	req.WriteBody(&pkt.Session{
		Account:   tk.Account,
		ChannelId: id,
		GateId:    h.ServiceID,
		App:       tk.App,
		RemoteIP:  getIP(conn.RemoteAddr().String()),
	})

	// 7. 把login转发给Login服务
	if err := container.Forward(wire.SNLogin, req); err != nil {
		return "", err
	}

	return id, nil
}

func (h *Handler) Receive(agent kim.Agent, payload []byte) {
	buf := bytes.NewBuffer(payload)

	packet, err := pkt.Read(buf)
	if err != nil {
		log.Error(err)
		return
	}

	// 如果是 BasicPkt，就处理心跳包
	if basicPkt, ok := packet.(*pkt.BasicPkt); ok {
		if basicPkt.Code == pkt.CodePing {
			_ = agent.Push(pkt.Marshal(&pkt.BasicPkt{Code: pkt.CodePong}))
		}
		return
	}

	// 如果是 LogicPkt，就转发给逻辑服务处理
	if logicPkt, ok := packet.(*pkt.LogicPkt); ok {
		logicPkt.ChannelId = agent.ID()

		err := container.Forward(logicPkt.ServiceName(), logicPkt)
		if err != nil {
			logger.WithFields(logger.Fields{
				"module": "handler",
				"id":     agent.ID(),
				"cmd":    logicPkt.Command,
				"dest":   logicPkt.Dest,
			}).Error(err)
		}
	}
}

func (h *Handler) Disconnect(id string) error {
	log.Infof("disconnect %s", id)

	logout := pkt.New(wire.CommandLoginSignOut, pkt.WithChannel(id))
	if err := container.Forward(wire.SNLogin, logout); err != nil {
		logger.WithFields(logger.Fields{
			"module": "handler",
			"id":     id,
		}).Error(err)
	}

	return nil
}

var ipExp = regexp.MustCompile(string("\\:[0-9]+$"))

func getIP(remoteAddr string) string {
	if remoteAddr == "" {
		return ""
	}

	return ipExp.ReplaceAllString(remoteAddr, "")
}

func generateChannelID(serviceID, account string) string {
	return fmt.Sprintf("%s_%s_%d", serviceID, account, wire.Seq.Next())
}
