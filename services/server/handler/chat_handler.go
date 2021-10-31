package handler

import (
	"errors"
	"time"

	"jinv/kim"
	"jinv/kim/services/server/service"
	"jinv/kim/wire/pkt"
	"jinv/kim/wire/rpc"
)

var ErrDoDestination = errors.New("desc is empty")

type ChatHandler struct {
	msgService   service.Message
	groupService service.Group
}

func NewChatHandler(message service.Message, group service.Group) *ChatHandler {
	return &ChatHandler{
		msgService:   message,
		groupService: group,
	}
}

func (h *ChatHandler) DoUserTalk(ctx kim.Context) {
	// 1. 校验接收人是否存在
	if ctx.Header().GetDest() == "" {
		_ = ctx.RespWithError(pkt.Status_NoDestination, ErrDoDestination)
		return
	}

	// 2. 解包
	var req pkt.MessageReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	// 3. 获取接收人的位置信息（所在的网关及ChannelId）
	receiver := ctx.Header().GetDest()
	loc, err := ctx.GetLocation(receiver, "")
	if err != nil && err != kim.ErrSessionNil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// 4. 保存离线消息
	sendTime := time.Now().UnixNano()
	resp, err := h.msgService.InsertUser(ctx.Session().GetApp(), &rpc.InsertMessageReq{
		Sender:   ctx.Session().GetAccount(),
		Dest:     receiver,
		SendTime: sendTime,
		Message: &rpc.Message{
			Type:  req.GetType(),
			Body:  req.GetBody(),
			Extra: req.GetExtra(),
		},
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	msgId := resp.MessageId

	// 5. 如果接收方在线，就推送一条消息过去 todo
	if loc != nil {
		err := ctx.Dispatch(&pkt.MessagePush{
			MessageId: msgId,
			Type:      req.GetType(),
			Body:      req.GetBody(),
			Extra:     req.GetExtra(),
			Sender:    ctx.Session().GetAccount(),
			SendTime:  sendTime,
		}, loc)
		if err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}

	// 6. 返回一条resp消息
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageResp{
		MessageId: msgId,
		SendTime:  sendTime,
	})
}

func (h *ChatHandler) DoGroupTalk(ctx kim.Context) {
	// 1. 校验接收人是否存在
	if ctx.Header().GetDest() == "" {
		_ = ctx.RespWithError(pkt.Status_NoDestination, ErrDoDestination)
	}

	// 2. 解包
	var req pkt.MessageReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	// 3. 群聊中的 dest 就不再是 user account, 而是群ID
	group := ctx.Header().GetDest()

	sendTime := time.Now().UnixNano()

	// 4. 保存离线消息
	resp, err := h.msgService.InsertGroup(ctx.Session().GetApp(), &rpc.InsertMessageReq{
		Sender:   ctx.Session().GetAccount(),
		Dest:     group,
		SendTime: sendTime,
		Message: &rpc.Message{
			Type:  req.GetType(),
			Body:  req.GetBody(),
			Extra: req.GetExtra(),
		},
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// 5. 读取成员列表
	membersResp, err := h.groupService.Members(ctx.Session().GetApp(), &rpc.GroupMembersReq{
		GroupId: group,
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	var members = make([]string, len(membersResp.Users))
	for i, user := range membersResp.Users {
		members[i] = user.Account
	}

	// 6. 批量寻址（群成员）
	locs, err := ctx.GetLocations(members...)
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// 7. 批量推送消息给群成员
	if len(locs) > 0 {
		err = ctx.Dispatch(&pkt.MessagePush{
			MessageId: resp.MessageId,
			Type:      req.GetType(),
			Body:      req.GetBody(),
			Extra:     req.GetExtra(),
			Sender:    ctx.Session().GetAccount(),
			SendTime:  sendTime,
		}, locs...)
		if err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}

	// 8. 返回一条 resp 消息
	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageResp{
		MessageId: resp.MessageId,
		SendTime:  sendTime,
	})
}

func (h *ChatHandler) DoTalkAck(ctx kim.Context) {
	var req pkt.MessageAckReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	err := h.msgService.SetAck(ctx.Session().GetApp(), &rpc.AckMessageReq{
		Account:   ctx.Session().GetAccount(),
		MessageId: req.GetMessageId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}
