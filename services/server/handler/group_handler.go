package handler

import (
	"jinv/kim"
	"jinv/kim/services/server/service"
	"jinv/kim/wire/pkt"
	"jinv/kim/wire/rpc"
)

type GroupHandler struct {
	groupService service.Group
}

func NewGroupHandler(groupService service.Group) *GroupHandler {
	return &GroupHandler{
		groupService: groupService,
	}
}

func (h *GroupHandler) DoCreate(ctx kim.Context) {
	var req pkt.GroupCreateReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	// 创建群
	resp, err := h.groupService.Create(ctx.Session().GetApp(), &rpc.CreateGroupReq{
		Name:         req.GetName(),
		Avatar:       req.GetAvatar(),
		Introduction: req.GetIntroduction(),
		Owner:        req.GetOwner(),
		Members:      req.GetMembers(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// 创建群后的消息通知 todo
	locs, err := ctx.GetLocations(req.GetMembers()...)
	if len(locs) > 0 {
		err = ctx.Dispatch(&pkt.GroupCreateNotify{
			GroupId: resp.GroupId,
			Members: req.GetMembers(),
		}, locs...)
		if err != nil {
			_ = ctx.RespWithError(pkt.Status_SystemException, err)
			return
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.GroupCreateResp{
		GroupId: resp.GroupId,
	})
}

func (h *GroupHandler) DoJoin(ctx kim.Context) {
	var req pkt.GroupJoinReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	err := h.groupService.Join(ctx.Session().GetApp(), &rpc.JoinGroupReq{
		Account: req.GetAccount(),
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	// todo 写Nil的情况
	// todo 加入群之后，应该也要通知在线的全员吧

	_ = ctx.Resp(pkt.Status_Success, nil)
}

func (h *GroupHandler) DoQuit(ctx kim.Context) {
	var req pkt.GroupQuitReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	err := h.groupService.Quit(ctx.Session().GetApp(), &rpc.QuitGroupReq{
		Account: req.GetAccount(),
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	_ = ctx.Resp(pkt.Status_Success, nil)
}

func (h *GroupHandler) DoDetail(ctx kim.Context) {
	var req pkt.GroupGetReq
	if err := ctx.ReadBody(&req); err != nil {
		_ = ctx.RespWithError(pkt.Status_InvalidPacketBody, err)
		return
	}

	respGroup, err := h.groupService.Detail(ctx.Session().GetApp(), &rpc.GetGroupReq{
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	respMembers, err := h.groupService.Members(ctx.Session().GetApp(), &rpc.GroupMembersReq{
		GroupId: req.GetGroupId(),
	})
	if err != nil {
		_ = ctx.RespWithError(pkt.Status_SystemException, err)
		return
	}

	members := make([]*pkt.Member, len(respMembers.GetUsers()))
	for i, member := range respMembers.GetUsers() {
		members[i] = &pkt.Member{
			Account:  member.GetAccount(),
			Alias:    member.GetAlias(),
			Avatar:   member.GetAvatar(),
			JoinTime: member.GetJoinTime(),
		}
	}

	_ = ctx.Resp(pkt.Status_Success, &pkt.GroupGetResp{
		Id:           respGroup.Id,
		Name:         respGroup.Name,
		Avatar:       respGroup.Avatar,
		Introduction: respGroup.Introduction,
		Owner:        respGroup.Owner,
		CreatedAt:    respGroup.CreatedAt,
		Member:       members,
	})
}
