package handler

import (
	"errors"

	"jinv/kim"
	"jinv/kim/services/server/service"
	"jinv/kim/wire/pkt"
	"jinv/kim/wire/rpc"
)

type OfflineHandler struct {
	messageService service.Message
}

func NewOfflineHandler(messageService service.Message) *OfflineHandler {
	return &OfflineHandler{
		messageService: messageService,
	}
}

func (h *OfflineHandler) DoSyncIndex(ctx kim.Context) {
	var req pkt.MessageIndexReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	resp, err := h.messageService.GetMessageIndex(ctx.Session().GetApp(), &rpc.GetOfflineMessageIndexReq{
		Account:   ctx.Session().GetAccount(),
		MessageId: req.GetMessageId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	indexes := make([]*pkt.MessageIndex, len(resp.GetList()))
	for i, index := range resp.GetList() {
		indexes[i] = &pkt.MessageIndex{
			MessageId: index.MessageId,
			Direction: index.Direction,
			SendTime:  index.SendTime,
			AccountB:  index.AccountB,
			Group:     index.Group,
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageIndexResp{
		Indexes: indexes,
	})
}

func (h *OfflineHandler) DoSyncContent(ctx kim.Context) {
	var req pkt.MessageContentReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	if len(req.GetMessageIds()) == 0 {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, errors.New("Empty MessageIds"))
		return
	}

	resp, err := h.messageService.GetMessageContent(ctx.Session().GetApp(), &rpc.GetOfflineMessageContentReq{
		MessageIds: req.GetMessageIds(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	contents := make([]*pkt.MessageContent, len(resp.GetList()))
	for i, content := range resp.GetList() {
		contents[i] = &pkt.MessageContent{
			MessageId: content.Id,
			Type:      content.Type,
			Body:      content.Body,
			Extra:     content.Extra,
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.MessageContentResp{
		Contents: contents,
	})
}
