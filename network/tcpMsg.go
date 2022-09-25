package network

import (
	"errors"
	"io"
	"unsafe"
)

type CSMsgHead struct {
	MsgLen uint32 // 消息长度
	Echo   uint32 // 用于校验包的对应关系，递增
	Cmd    uint32 // 消息ID
}

type SSMsgHead struct {
	MsgLen  uint32 // 消息长度
	Echo    uint32 // 用于校验包的对应关系，递增
	Cmd     uint32 // 消息ID
	AgentID uint32 // 客户端连接的代理ID
}

type MsgHeadByte struct {
	addr uintptr
	len  int
	cap  int
}

type MsgParser struct {
	minMsgLen    uint32
	maxMsgLen    uint32
	littleEndian bool
}

func NewMsgParser() *MsgParser {
	msgParser := new(MsgParser)
	msgParser.minMsgLen = 1
	msgParser.maxMsgLen = 4096
	msgParser.littleEndian = false

	return msgParser
}

func (msg *MsgParser) SetMsgLen(minMsgLen uint32, maxMsgLen uint32) {
	if minMsgLen != 0 {
		msg.minMsgLen = minMsgLen
	}

	if maxMsgLen != 0 {
		msg.maxMsgLen = maxMsgLen
	}
}

func (msg *MsgParser) SetByteOrder(littleEndian bool) {
	msg.littleEndian = littleEndian
}

func (msg *MsgParser) Read(conn *TCPConn, args ...[]byte) ([]byte, error) {
	CSMsgHeadLen := int(unsafe.Sizeof(CSMsgHead{}))
	msgHeadBuf := make([]byte, CSMsgHeadLen)
	if _, err := io.ReadFull(conn, msgHeadBuf); err != nil {
		return nil, err
	}

	CSMsgHead := *(**CSMsgHead)(unsafe.Pointer(&msgHeadBuf))

	if CSMsgHead.MsgLen > msg.maxMsgLen {
		return nil, errors.New("message too long")
	} else if CSMsgHead.MsgLen < msg.minMsgLen {
		return nil, errors.New("message too short")
	}

	msgData := make([]byte, CSMsgHead.MsgLen)
	if _, err := io.ReadFull(conn, msgData); err != nil {
		return nil, err
	}
	msgBuf := PackSSMsgHead(CSMsgHead, msgData)
	return msgBuf, nil
}

func (msg *MsgParser) Write(conn *TCPConn, args ...[]byte) error {
	var msgLen uint32
	for i := 0; i < len(args); i++ {
		msgLen += uint32(len(args[i]))
	}

	if msgLen > msg.maxMsgLen {
		return errors.New("message too long")
	} else if msgLen < msg.minMsgLen {
		return errors.New("message too short")
	}

	// mssage := make([]byte, uint32(msg.msgHeadLen)+msgLen)

	// switch msg.msgHeadLen {
	// case 1:
	// 	mssage[0] = byte(msgLen)
	// case 2:
	// 	if msg.littleEndian {
	// 		binary.LittleEndian.PutUint16(mssage, uint16(msgLen))
	// 	} else {
	// 		binary.BigEndian.PutUint16(mssage, uint16(msgLen))
	// 	}
	// case 4:
	// 	if msg.littleEndian {
	// 		binary.LittleEndian.PutUint32(mssage, uint32(msgLen))
	// 	} else {
	// 		binary.BigEndian.PutUint32(mssage, uint32(msgLen))
	// 	}
	// }

	// l := msg.msgHeadLen
	// for i := 0; i < len(args); i++ {
	// 	copy(mssage[l:], args[i])
	// 	l += len(args[i])
	// }

	// conn.Write(mssage)

	return nil
}

func PackSSMsgHead(CSMsgHead *CSMsgHead, msgData []byte) []byte {
	SSMsgHeadLen := int(unsafe.Sizeof(SSMsgHead{}))
	msgBuf := make([]byte, SSMsgHeadLen+len(msgData))

	SSMsgHead := &SSMsgHead{
		MsgLen:  CSMsgHead.MsgLen,
		Echo:    CSMsgHead.Echo,
		Cmd:     CSMsgHead.Cmd,
		AgentID: 1,
	}

	Len := unsafe.Sizeof(*SSMsgHead)
	msgHeadByte := &MsgHeadByte{
		addr: uintptr(unsafe.Pointer(SSMsgHead)),
		cap:  int(Len),
		len:  int(Len),
	}

	msgHeadBuf := *(*[]byte)(unsafe.Pointer(msgHeadByte))

	copy(msgBuf[:Len], msgHeadBuf)
	copy(msgBuf[Len:], msgData)

	return msgBuf
}
