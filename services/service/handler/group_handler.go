package handler

import (
	"errors"

	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"jinv/kim/services/service/database"
	"jinv/kim/wire/rpc"
)

func (h *ServiceHandler) GroupCreate(c iris.Context) {
	var req rpc.CreateGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	app := c.Params().Get("app") // todo 这里为什么从这里获取

	groupId := h.IdGen.Next()

	group := &database.Group{
		Model: database.Model{
			ID: groupId.Int64(),
		},
		App:          app,
		Group:        groupId.Base36(),
		Name:         req.Name,
		Avatar:       req.Avatar,
		Owner:        req.Owner,
		Introduction: req.Introduction,
	}

	members := make([]database.GroupMember, len(req.Members))

	for i, member := range req.Members {
		members[i] = database.GroupMember{
			Model: database.Model{
				ID: h.IdGen.Next().Int64(),
			},
			Account: member,
			Group:   groupId.Base36(),
		}
	}

	err := h.BaseDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(group).Error; err != nil {
			return err
		}

		return tx.Create(members).Error
	})
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.CreateGroupResp{
		GroupId: groupId.Base36(),
	})
}

func (h *ServiceHandler) GroupJoin(c iris.Context) {
	var req rpc.JoinGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	groupMember := &database.GroupMember{
		Model: database.Model{
			ID: h.IdGen.Next().Int64(),
		},
		Account: req.Account,
		Group:   req.GroupId,
	}

	if err := h.BaseDb.Create(groupMember).Error; err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupQuit(c iris.Context) {
	var req rpc.QuitGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	groupMember := &database.GroupMember{
		Account: req.Account,
		Group:   req.GroupId,
	}

	if err := h.BaseDb.Delete(&database.GroupMember{}, groupMember).Error; err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func (h *ServiceHandler) GroupMembers(c iris.Context) {
	groupId := c.Params().Get("id")
	if groupId == "" {
		c.StopWithError(iris.StatusBadRequest, errors.New("group is null"))
		return
	}

	var members []database.GroupMember
	err := h.BaseDb.Order("updated_at asc").Find(&members, database.GroupMember{Group: groupId}).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	var users = make([]*rpc.Member, len(members))
	for i, member := range members {
		users[i] = &rpc.Member{
			Account:  member.Account,
			Alias:    member.Alias,
			JoinTime: member.CreatedAt.Unix(),
		}
	}

	_, _ = c.Negotiate(&rpc.GroupMembersResp{
		Users: users,
	})
}

func (h *ServiceHandler) GroupGet(c iris.Context) {
	var req rpc.GetGroupReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	id, err := h.IdGen.ParseBase36(req.GroupId)
	if err != nil {
		c.StopWithError(iris.StatusBadRequest, errors.New("group is invalid :"+req.GroupId))
	}

	var group database.Group

	if err := h.BaseDb.First(&group, id.Int64()).Error; err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.GetGroupResp{
		Id:           req.GroupId,
		Name:         group.Name,
		Avatar:       group.Avatar,
		Introduction: group.Introduction,
		Owner:        group.Owner,
		CreatedAt:    group.CreatedAt.Unix(),
	})
}
