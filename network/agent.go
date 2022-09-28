package network

import (
	"gateServer/log"
	"gateServer/msg"
	"gateServer/nsqer"
	"net"
	"time"
)

const (
	_ int32 = iota
	STATUS_START
	STATUS_HANDSHAKE
	STATUS_WORKING
	STATUS_CLOSED
)

const (
	CS_MSG_HEAD_LEN  = 12
	RECV_BUF_MAX_LEN = 10240
)

type agent struct {
	// regular agent member
	ID      int64         // agent unique id
	UID     uint64        // binding user id
	conn    net.Conn      // low-level conn fd
	lastMid uint64        // last message id
	state   int32         // current agent state
	chDie   chan struct{} // wait for close
	chSend  chan []byte   // push message queue
	lastAt  int64         // last heartbeat unix time stamp

	recvBuf    [RECV_BUF_MAX_LEN]byte
	recvBufLen uint32
}

// Create new agent instance
func newAgent(conn net.Conn) *agent {
	a := &agent{
		ID:         Connections.SessionID(),
		conn:       conn,
		state:      STATUS_START,
		chDie:      make(chan struct{}),
		lastAt:     time.Now().Unix(),
		chSend:     make(chan []byte, 10),
		recvBufLen: 0,
	}

	Connections.Increment()
	return a
}

func (pThis *agent) Run() {
	go pThis.write()
	pThis.read()
}

func (pThis *agent) read() {
	for {
		recvBuf := make([]byte, RECV_BUF_MAX_LEN-pThis.recvBufLen)
		recvLen, err := pThis.conn.Read(recvBuf)
		if err != nil {
			log.Log.WithField("Err", err).Debug("read message")
			break
		}
		log.Log.WithField("recvlen", recvLen).Debug("recv message")
		copy(pThis.recvBuf[pThis.recvBufLen:], recvBuf[:recvLen])
		pThis.recvBufLen += uint32(recvLen)
		// 处理消息
		pThis.handleMsg()
		log.Log.WithField("msgCnt", pThis.lastMid).Debug()
	}
}

func (pThis *agent) handleMsg() {
	for pThis.recvBufLen > CS_MSG_HEAD_LEN {
		head := msg.UnPackCSMsgHead(pThis.recvBuf[:CS_MSG_HEAD_LEN])
		if head.MsgLen > pThis.recvBufLen {
			return
		}

		if pThis.state == STATUS_START {
			pThis.state = STATUS_HANDSHAKE
		}
		if pThis.state == STATUS_WORKING || head.Cmd == 1 {

		}
		msgBodyBuf := pThis.recvBuf[CS_MSG_HEAD_LEN : CS_MSG_HEAD_LEN+head.MsgLen]
		msgBuf := msg.PackSSMsgHead(head, msgBodyBuf, pThis.ID)
		nsqer.NsqProducer.Publish("client", msgBuf)
		pThis.lastMid++
		copy(pThis.recvBuf[:], pThis.recvBuf[:CS_MSG_HEAD_LEN+head.MsgLen])
		pThis.recvBufLen -= CS_MSG_HEAD_LEN + head.MsgLen
	}
}

func (pThis *agent) write() {
	for b := range pThis.chSend {
		if b == nil {
			break
		}
		log.Log.WithField("sendLen", len(b)).Debug()
		_, err := pThis.conn.Write(b)

		if err != nil {
			break
		}
	}
	pThis.OnClose()
}

func (pThis *agent) PushSendMsg(msgBuf []byte) {
	pThis.chSend <- msgBuf
}

func (pThis *agent) OnClose() {
	pThis.conn.Close()
	Connections.Decrement()
}
