package gate

import (
	"gateServer/log"
	"gateServer/msg"
	"gateServer/network"
	"unsafe"

	"github.com/nsqio/go-nsq"
)

type ClientMsgConsumer struct {
	Server *network.TCPServer
}

func (pThis *ClientMsgConsumer) HandleMessage(message *nsq.Message) error {
	// 解析消息头
	if len(message.Body) > msg.SS_MSG_HEAD_LEN {
		msgHeadBuf := message.Body[:msg.SS_MSG_HEAD_LEN]
		msgHead := *(**msg.SSMsgHead)(unsafe.Pointer(&msgHeadBuf))
		agent := pThis.Server.GetAgent(msgHead.AgentID)
		if agent != nil {
			msgBuf := msg.PackCSMsgHead(msgHead, message.Body[msg.SS_MSG_HEAD_LEN:])

			log.Log.WithField("recvLen", len(msgBuf)).Debug()
			agent.PushSendMsg(msgBuf)
		}
	}

	message.Finish()
	return nil
}
