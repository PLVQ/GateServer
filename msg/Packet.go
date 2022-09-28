package msg

import "unsafe"

func UnPackCSMsgHead(headBuf []byte) *CSMsgHead {
	head := *(**CSMsgHead)(unsafe.Pointer(&headBuf))
	return head
}

func PackCSMsgHead(msgHead *SSMsgHead, msgData []byte) []byte {
	msgLen := CS_MSG_HEAD_LEN + len(msgData)
	msgBuf := make([]byte, msgLen)

	CSMsgHead := &CSMsgHead{
		MsgLen: msgHead.MsgLen,
		Echo:   msgHead.Echo,
		Cmd:    msgHead.Cmd,
	}

	msgHeadByte := &MsgHeadByte{
		addr: uintptr(unsafe.Pointer(CSMsgHead)),
		cap:  CS_MSG_HEAD_LEN,
		len:  CS_MSG_HEAD_LEN,
	}

	msgHeadBuf := *(*[]byte)(unsafe.Pointer(msgHeadByte))

	copy(msgBuf[:CS_MSG_HEAD_LEN], msgHeadBuf)
	copy(msgBuf[CS_MSG_HEAD_LEN:], msgData)

	return msgBuf
}

func PackSSMsgHead(msgHead *CSMsgHead, msgData []byte, agentID int64) []byte {
	msgLen := SS_MSG_HEAD_LEN + len(msgData)
	msgBuf := make([]byte, msgLen)

	SSMsgHead := &SSMsgHead{
		MsgLen:  msgHead.MsgLen,
		Echo:    msgHead.Echo,
		Cmd:     msgHead.Cmd,
		AgentID: agentID,
	}

	msgHeadByte := &MsgHeadByte{
		addr: uintptr(unsafe.Pointer(SSMsgHead)),
		cap:  SS_MSG_HEAD_LEN,
		len:  SS_MSG_HEAD_LEN,
	}

	msgHeadBuf := *(*[]byte)(unsafe.Pointer(msgHeadByte))

	copy(msgBuf[:SS_MSG_HEAD_LEN], msgHeadBuf)
	copy(msgBuf[SS_MSG_HEAD_LEN:], msgData)

	return msgBuf
}
