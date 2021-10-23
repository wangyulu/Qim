package handler

import (
	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/wire/pkt"
)

type LoginHandler struct {
}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

func (h *LoginHandler) DoSysLogin(ctx kim.Context) {
	// 1. 序列化
	var session pkt.Session

	if err := ctx.ReadBody(&session); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)

		return
	}

	logger.WithFields(logger.Fields{
		"Func":      "Login",
		"ChannelId": session.GetChannelId(),
		"Account":   session.GetAccount(),
		"RemoteIP":  session.GetRemoteIP(),
	}).Info("do login")

	// 2. 检查当前账号是否已经登录在其它地方
	old, err := ctx.GetLocation(session.Account, "")
	if err != nil && err != kim.ErrSessionNil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
	}

	if old != nil {
		// 3. 通知这个用户下线
		_ = ctx.Dispatch(&pkt.KichoutNotify{ChannelId: old.ChannelId}, old)
	}

	// 4. 添加到会话管理器中
	if err := ctx.Add(&session); err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)

		return
	}

	// 5. 返回一个登录成功的消息
	var resp = &pkt.LoginResp{
		ChannelId: session.ChannelId,
	}

	_ = ctx.Resp(pkt.Status_Success, resp)
}

func (h *LoginHandler) DoSysLogout(ctx kim.Context) {
	logger.WithFields(logger.Fields{
		"Func":      "Logout",
		"ChannelId": ctx.Session().GetChannelId(),
		"Account":   ctx.Session().GetAccount(),
	}).Info("do logout")

	err := ctx.Delete(ctx.Session().GetAccount(), ctx.Session().GetChannelId())
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)

		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}
