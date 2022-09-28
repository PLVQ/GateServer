package msg

import "unsafe"

type CSMsgHead struct {
	MsgLen uint32 // 消息长度
	Echo   uint32 // 用于校验包的对应关系，递增
	Cmd    uint32 // 消息ID
}

type SSMsgHead struct {
	MsgLen  uint32 // 消息长度
	Echo    uint32 // 用于校验包的对应关系，递增
	Cmd     uint32 // 消息ID
	AgentID int64  // 客户端连接的代理ID
}

type MsgHeadByte struct {
	addr uintptr
	len  int
	cap  int
}

const (
	CS_MSG_HEAD_LEN = int(unsafe.Sizeof(CSMsgHead{}))
	SS_MSG_HEAD_LEN = int(unsafe.Sizeof(SSMsgHead{}))
)
