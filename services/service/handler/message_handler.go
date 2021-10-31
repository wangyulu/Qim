package handler

import (
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
	"jinv/kim/services/service/database"
	"jinv/kim/wire"
	"jinv/kim/wire/rpc"
)

type ServiceHandler struct {
	BaseDb    *gorm.DB
	MessageDb *gorm.DB
	Cache     *redis.Client
	IdGen     *database.IDGenerator
}

func (h *ServiceHandler) InsertUserMessage(c iris.Context) {
	var req rpc.InsertMessageReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	messageId := h.IdGen.Next().Int64()
	messageContent := database.MessageContent{
		ID:       messageId,
		Type:     byte(req.Message.Type),
		Body:     req.Message.Body,
		Extra:    req.Message.Extra,
		SendTime: req.SendTime,
	}

	// 扩展写
	indexes := make([]database.MessageIndex, 2)
	indexes[0] = database.MessageIndex{
		ID:        h.IdGen.Next().Int64(),
		MessageID: messageId,
		AccountA:  req.Dest,
		AccountB:  req.Sender,
		Direction: 0,
		SendTime:  req.SendTime,
	}

	indexes[1] = database.MessageIndex{
		ID:        h.IdGen.Next().Int64(),
		MessageID: messageId,
		AccountA:  req.Sender,
		AccountB:  req.Dest,
		Direction: 1,
		SendTime:  req.SendTime,
	}

	err := h.MessageDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&messageContent).Error; err != nil {
			return err
		}

		return tx.Create(&indexes).Error
	})
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.InsertMessageResp{
		MessageId: messageId,
	})
}

func (h *ServiceHandler) InsertGroupMessage(c iris.Context) {
	var req rpc.InsertMessageReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	var members []database.GroupMember
	err := h.BaseDb.Where(&database.GroupMember{Group: req.Dest}).Find(&members).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	// 扩散写
	messageId := h.IdGen.Next().Int64()
	messageContent := database.MessageContent{
		ID:       messageId,
		Type:     byte(req.Message.Type),
		Body:     req.Message.Body,
		Extra:    req.Message.Extra,
		SendTime: req.SendTime,
	}

	var indexes = make([]database.MessageIndex, len(members))
	for i, member := range members {
		indexes[i] = database.MessageIndex{
			ID:        h.IdGen.Next().Int64(),
			MessageID: messageId,
			AccountA:  member.Account,
			AccountB:  req.Sender,
			Direction: 0,
			Group:     member.Group,
			SendTime:  req.SendTime,
		}

		if req.Sender == member.Account {
			indexes[i].Direction = 1
		}
	}

	err = h.MessageDb.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&messageContent).Error; err != nil {
			return err
		}

		return tx.Create(&indexes).Error
	})
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.InsertMessageResp{
		MessageId: messageId,
	})
}

func (h *ServiceHandler) MessageAck(c iris.Context) {
	var req rpc.AckMessageReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	if err := setMessageAck(h.Cache, req.Account, req.MessageId); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}
}

func setMessageAck(cache *redis.Client, account string, msgId int64) error {
	if msgId == 0 {
		return nil
	}

	key := database.KeyMessageAckIndex(account)

	return cache.Set(key, msgId, wire.OfflineReadIndexExpiresIn).Err()
}

func (h *ServiceHandler) GetOfflineMessageIndex(c iris.Context) {
	var req rpc.GetOfflineMessageIndexReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	msgId := req.MessageId
	start, err := h.getSendTime(req.Account, msgId)
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	var indexes []*rpc.MessageIndex
	tx := h.MessageDb.Model(&database.MessageIndex{}).Select("account_b", "direction", "message_id", "group", "send_time")
	tx = tx.Where("account_a =? and send_time > ? and direction = ?", req.Account, start, 0)
	err = tx.Order("send_time asc").Limit(wire.OfflineSyncIndexCount).Find(&indexes).Error
	if err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	if err := setMessageAck(h.Cache, req.Account, msgId); err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.GetOfflineMessageIndexResp{
		List: indexes,
	})
}

func (h *ServiceHandler) getSendTime(account string, msgId int64) (int64, error) {
	// 1. 冷启动情况，从服务端拉取消息索引
	if msgId == 0 {
		key := database.KeyMessageAckIndex(account)
		msgId, _ = h.Cache.Get(key).Int64()
	}

	var start int64
	if msgId > 0 {
		// 2. 根据消息ID获取此消息的发送时间
		var content database.MessageContent
		err := h.MessageDb.Select("send_time").First(&content, msgId).Error
		if err != nil {
			// 3. 如果此条消息不存在，返回最近一天
			start = time.Now().AddDate(0, 0, -1).UnixNano()
		} else {
			start = content.SendTime
		}
	}

	// 4. 返回默认的离线消息过期时间
	earliestKeepTime := time.Now().AddDate(0, 0, -1*wire.OfflineMessageExpiresIn).UnixNano()
	if start == 0 || start < earliestKeepTime {
		start = earliestKeepTime
	}

	return start, nil
}

func (h *ServiceHandler) GetOfflineMessageContent(c iris.Context) {
	var req rpc.GetOfflineMessageContentReq
	if err := c.ReadBody(&req); err != nil {
		c.StopWithError(iris.StatusBadRequest, err)
		return
	}

	messageLen := len(req.MessageIds)
	if messageLen > wire.MessageMaxCountPerPage {
		c.StopWithText(iris.StatusBadRequest, "too many message_ids")
		return
	}

	var contents []*rpc.Message
	if err := h.MessageDb.Model(&database.MessageContent{}).Where(req.MessageIds).Find(&contents).Error; err != nil {
		c.StopWithError(iris.StatusInternalServerError, err)
		return
	}

	_, _ = c.Negotiate(&rpc.GetOfflineMessageContentResp{
		List: contents,
	})
}
