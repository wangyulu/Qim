package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"jinv/kim/wire"
	"jinv/kim/wire/rpc"
)

var messageService = NewMessageService("http://localhost:8080")

func Test_UserMessage(t *testing.T) {
	message := &rpc.Message{
		Type: wire.MessageTypeText,
		Body: "hello world",
	}

	sender := fmt.Sprintf("%d_test1", time.Now().UnixNano())
	dest := fmt.Sprintf("%d_test2", time.Now().UnixNano())

	resp1, err := messageService.InsertUser(app, &rpc.InsertMessageReq{
		Sender:   sender,
		Dest:     dest,
		SendTime: time.Now().UnixNano(),
		Message:  message,
	})
	assert.Nil(t, err)

	resp2, err := messageService.GetMessageIndex(app, &rpc.GetOfflineMessageIndexReq{
		Account: dest,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp2.GetList()))
	assert.Equal(t, sender, resp2.GetList()[0].AccountB)

	resp3, err := messageService.GetMessageContent(app, &rpc.GetOfflineMessageContentReq{
		MessageIds: []int64{resp1.GetMessageId()},
	})
	assert.Nil(t, err)
	assert.Equal(t, message.Body, resp3.GetList()[0].Body)
	assert.Equal(t, resp1.MessageId, resp3.GetList()[0].Id)

	resp4, err := messageService.GetMessageIndex(app, &rpc.GetOfflineMessageIndexReq{
		Account:   dest,
		MessageId: resp1.MessageId,
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp4.GetList()))

	resp5, err := messageService.GetMessageIndex(app, &rpc.GetOfflineMessageIndexReq{
		Account: dest,
	})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp5.GetList()))
}

func Test_GroupMessage(t *testing.T) {
	test1 := fmt.Sprintf("%d_test1", time.Now().UnixNano())
	test2 := fmt.Sprintf("%d_test2", time.Now().UnixNano())
	test3 := fmt.Sprintf("%d_test3", time.Now().UnixNano())

	resp1, err := groupService.Create(app, &rpc.CreateGroupReq{
		Name:    "test",
		Owner:   test1,
		Members: []string{test1, test2, test3},
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, resp1.GroupId)
	t.Log(resp1.GroupId)

	message := &rpc.Message{
		Type: wire.MessageTypeText,
		Body: "hello world",
	}

	_, err = messageService.InsertGroup(app, &rpc.InsertMessageReq{
		Sender:   test1,
		Dest:     resp1.GroupId,
		SendTime: time.Now().UnixNano(),
		Message:  message,
	})
	assert.Nil(t, err)

	resp2, err := messageService.GetMessageIndex(app, &rpc.GetOfflineMessageIndexReq{
		Account: test1,
	})
	assert.Nil(t, err)
	assert.NotEmpty(t, resp2)
	// assert.Equal(t, 0, len(resp2.GetList())) todo

	resp3, err := messageService.GetMessageIndex(app, &rpc.GetOfflineMessageIndexReq{
		Account: test2,
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(resp3.GetList()))
	assert.Equal(t, int32(0), resp3.GetList()[0].Direction)
}
