package unittest

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"jinv/kim"
	"jinv/kim/wire"
	"jinv/kim/wire/pkt"
)

func Test_UserTalk(t *testing.T) {
	// 登录用户：test1
	cli1, err := login("test1")
	assert.Nil(t, err)

	// 登录用户：test2
	cli2, err := login("test2")
	assert.Nil(t, err)

	// 构造消息包（test1 -> test2）
	packet := pkt.New(wire.CommandChatUserTalk, pkt.WithDest("test2"))
	packet.WriteBody(&pkt.MessageReq{
		Type: wire.MessageTypeText,
		Body: "hello world",
	})

	// 发送消息包（test1 -> test2）
	err = cli1.Send(pkt.Marshal(packet))
	assert.Nil(t, err)

	// 读取返回消息包
	frame, _ := cli1.Read()
	assert.Equal(t, kim.OpBinary, frame.GetOpCode())

	packet, err = pkt.MustReadLogicPkt(bytes.NewBuffer(frame.GetPayload()))
	assert.Nil(t, err)
	assert.Equal(t, pkt.Status_Success, packet.Header.Status)

	// 解包断言
	var resp pkt.MessageResp
	_ = packet.ReadBody(&resp)
	assert.Greater(t, resp.MessageId, int64(1000))
	assert.Greater(t, resp.SendTime, int64(1000))
	t.Log(&resp)

	// test2 读取 test1 推送过来的消息
	frame, err = cli2.Read()
	assert.Nil(t, err)

	packet, err = pkt.MustReadLogicPkt(bytes.NewBuffer(frame.GetPayload()))
	assert.Nil(t, err)

	var push pkt.MessagePush
	_ = packet.ReadBody(&push)

	assert.Equal(t, resp.MessageId, push.MessageId)
	assert.Equal(t, resp.SendTime, push.SendTime)
	assert.Equal(t, "hello world", push.Body)
	assert.Equal(t, int32(1), push.Type)
	t.Log(&push)
}

func Test_GroupTalk(t *testing.T) {
	// 1. 用户 Test1 登录
	cli1, err := login("test1")
	assert.Nil(t, err)

	// 2. 创建群
	packet := pkt.New(wire.CommandGroupCreate)
	packet.WriteBody(&pkt.GroupCreateReq{
		Name:    "group1",
		Owner:   "test1",
		Members: []string{"test1", "test2", "test3", "test4"},
	})
	err = cli1.Send(pkt.Marshal(packet))
	assert.Nil(t, err)

	// 3. 读取创建群返回信息
	ack, err := cli1.Read()
	assert.Nil(t, err)

	ackPacket, err := pkt.MustReadLogicPkt(bytes.NewBuffer(ack.GetPayload()))
	assert.Equal(t, pkt.Status_Success, ackPacket.GetStatus())
	assert.Equal(t, wire.CommandGroupCreate, ackPacket.GetCommand())

	// 4. 解包
	var createdResp pkt.GroupCreateResp
	err = ackPacket.ReadBody(&createdResp)
	assert.Nil(t, err)

	group := createdResp.GetGroupId()
	assert.NotEmpty(t, group)
	if group == "" {
		return
	}

	// 5. 群成员 test2、test3登录
	cli2, err := login("test2")
	assert.Nil(t, err)

	cli3, err := login("test3")
	assert.Nil(t, err)

	t1 := time.Now()

	// 6. 发送群消息 CommandChatGroupTalk
	groupTalkPacket := pkt.New(wire.CommandChatGroupTalk, pkt.WithDest(group))
	groupTalkPacket.WriteBody(&pkt.MessageReq{
		Type: wire.MessageTypeText,
		Body: "hello group",
	})
	err = cli1.Send(pkt.Marshal(groupTalkPacket))
	assert.Nil(t, err)

	// 7. 读取resp消息，确认消息发送成功
	ack, _ = cli1.Read()
	ackPacket, _ = pkt.MustReadLogicPkt(bytes.NewBuffer(ack.GetPayload()))
	assert.Equal(t, pkt.Status_Success, ackPacket.GetStatus())

	// 8. test2 读取消息
	notify1, _ := cli2.Read()
	notify1Packet, err := pkt.MustReadLogicPkt(bytes.NewBuffer(notify1.GetPayload()))
	assert.Equal(t, wire.CommandChatGroupTalk, notify1Packet.GetCommand())

	var notify pkt.MessagePush
	_ = notify1Packet.ReadBody(&notify)

	// 9. 检验消息内容
	assert.Equal(t, "hello group", notify.GetBody())
	assert.Equal(t, int32(wire.MessageTypeText), notify.GetType())
	assert.Empty(t, notify.GetExtra())
	assert.Greater(t, notify.SendTime, t1.UnixNano())
	assert.Greater(t, notify.MessageId, int64(10000))

	// 10. test3 读取消息
	notify2, _ := cli3.Read()
	notify2Packet, err := pkt.MustReadLogicPkt(bytes.NewBuffer(notify2.GetPayload()))
	_ = notify2Packet.ReadBody(&notify)

	assert.Equal(t, "hello group", notify.GetBody())

	t.Logf("cost %v", time.Since(t1))
}
