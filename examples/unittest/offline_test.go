package unittest

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"jinv/kim"
	"jinv/kim/logger"
	"jinv/kim/wire"
	"jinv/kim/wire/pkt"
)

func Test_Offline(t *testing.T) {
	// 发送者与接收者
	sender := fmt.Sprintf("u%d", time.Now().Unix())
	receiver := fmt.Sprintf("u%d", time.Now().Unix()+1)

	// 登录发送者
	cliSender, err := login(sender)
	assert.Nil(t, err)

	// 发送10条消息
	count := 10
	for i := 0; i < count; i++ {
		packet := pkt.New(wire.CommandChatUserTalk, pkt.WithDest(receiver))
		packet.WriteBody(&pkt.MessageReq{
			Type: wire.MessageTypeText,
			Body: "hello world",
		})

		err = cliSender.Send(pkt.Marshal(packet))
		if err != nil {
			logger.Error(err)
			return
		}

		// 这里必须要主动对发送的结果进行接收
		_, _ = cliSender.Read()
	}

	cliReceiver, err := login(receiver)
	assert.Nil(t, err)

	// 获取接收者的消息索引
	packet := pkt.New(wire.CommandOfflineIndex)
	packet.WriteBody(&pkt.MessageIndexReq{})
	err = cliReceiver.Send(pkt.Marshal(packet))
	assert.Nil(t, err)

	var resp1 pkt.MessageIndexResp
	err = Read(cliReceiver, &resp1)
	assert.Nil(t, err)

	assert.Equal(t, count, len(resp1.Indexes))
	assert.Equal(t, sender, resp1.Indexes[0].AccountB)
	assert.Equal(t, int32(0), resp1.Indexes[0].Direction)
	t.Log(resp1.Indexes)

	messageIds := make([]int64, len(resp1.Indexes))
	for i, index := range resp1.Indexes {
		messageIds[i] = index.MessageId
	}
	t.Log(messageIds)

	// 拉取下一页
	lastMessageId := messageIds[count-1]
	packet = pkt.New(wire.CommandOfflineIndex)
	packet.WriteBody(&pkt.MessageIndexReq{
		MessageId: lastMessageId,
	})
	err = cliReceiver.Send(pkt.Marshal(packet))
	assert.Nil(t, err)

	var resp2 pkt.MessageIndexResp
	err = Read(cliReceiver, &resp2)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(resp2.GetIndexes()))

	// 获取消息内容
	packet = pkt.New(wire.CommandOfflineContext)
	packet.WriteBody(&pkt.MessageContentReq{MessageIds: messageIds})
	err = cliReceiver.Send(pkt.Marshal(packet))
	assert.Nil(t, err)

	var resp3 pkt.MessageContentResp
	err = Read(cliReceiver, &resp3)
	assert.Nil(t, err)

	for _, content := range resp3.GetContents() {
		t.Log(content.GetBody(), content.MessageId)
	}

	assert.Equal(t, count, len(resp3.GetContents()))
	assert.Equal(t, "hello world", resp3.GetContents()[0].Body)
}

func Read(cli kim.Client, body proto.Message) error {
	frame, err := cli.Read()
	if err != nil {
		return err
	}

	packet, _ := pkt.MustReadLogicPkt(bytes.NewBuffer(frame.GetPayload()))
	if packet.Status != pkt.Status_Success {
		return fmt.Errorf("received status: %v", packet.Status)
	}

	return packet.ReadBody(body)
}
