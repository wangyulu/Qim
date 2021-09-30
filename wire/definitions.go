package wire

// Command defined data type between client and server
const (
	// login
	CommandLoginSignIn  = "login.signin"
	CommandLoginSignOut = "login.signout"

	// chat
	CommandChatUserTalk  = "chat.user.talk"  // 发单聊消息
	CommandChatGroupTalk = "chat.group.talk" // 发群聊消息
	CommandChatTalkAck   = "chat.talk.ack"   // ACK TODO 这是个啥？回复？

	// 离线
	CommandOfflineIndex   = "chat.offline.index"   // 下载索引
	CommandOfflineContext = "chat.offline.context" // 下载内容

	// 群管理
	CommandGroupCreate  = "chat.group.create"  // 群创建
	CommandGroupJoin    = "chat.group.join"    // 加入群
	CommandGroupQuit    = "chat.group.quit"    // 退出群
	CommandGroupMembers = "chat.group.members" // 群成员
	CommandGroupDetail  = "chat.group.detail"  // 群详情
)

// Meta Key of a packet
const (
	MetaDestServer   = "dest.server"
	MetaDestChannels = "dest.channels"
)

type Protocal string

// Protocol
const (
	ProtocolTCP       Protocal = "tcp"
	ProtocolWebsocket Protocal = "websocket"
)

// Service Name 定义统一的服务名
const (
	SNWGateway = "wgateway" // websocket 网关服务
	SNTGateway = "tgateway" // tcp 网关服务
	SNLogin    = "login"    // 登录服务
	SNChat     = "chat"     // 聊天服务
	SNService  = "service"  // rpc服务
)

type ServiceID string

type SessionID string

type Magic [4]byte

var (
	MagicLogicPkt = Magic{0xc3, 0x11, 0xa3, 0x65}
	MagicBasicPkt = Magic{0xc3, 0x15, 0xa7, 0x65}
)
