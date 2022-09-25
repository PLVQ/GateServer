package gate

import (
	"net"
	"reflect"
	"time"

	"gateServer/log"
	"gateServer/network"
	"gateServer/nsqproducer"

	"github.com/sirupsen/logrus"
)

type Gate struct {
	MaxConnNum int
	MaxMsgLen  uint32
	Processor  network.Processor

	// websocket
	WSAddr      string
	HTTPTimeout time.Duration
	CertFile    string
	KeyFile     string

	// tcp
	TCPAddr      string
	LenMsgLen    int
	LittleEndian bool
}

func (gate *Gate) Run(closeSig chan bool) {
	// var wsServer *network.WSServer
	// if gate.WSAddr != "" {
	// 	wsServer = new(network.WSServer)
	// 	wsServer.Addr = gate.WSAddr
	// 	wsServer.MaxConnNum = gate.MaxConnNum
	// 	wsServer.PendingWriteNum = gate.PendingWriteNum
	// 	wsServer.MaxMsgLen = gate.MaxMsgLen
	// 	wsServer.HTTPTimeout = gate.HTTPTimeout
	// 	wsServer.CertFile = gate.CertFile
	// 	wsServer.KeyFile = gate.KeyFile
	// 	wsServer.NewAgent = func(conn *network.WSConn) network.Agent {
	// 		a := &agent{conn: conn, gate: gate}
	// 		if gate.AgentChanRPC != nil {
	// 			gate.AgentChanRPC.Go("NewAgent", a)
	// 		}
	// 		return a
	// 	}
	// }

	var tcpServer *network.TCPServer
	if gate.TCPAddr != "" {
		tcpServer = new(network.TCPServer)
		tcpServer.Addr = gate.TCPAddr
		tcpServer.MaxConnNum = gate.MaxConnNum
		tcpServer.LenMsgLen = gate.LenMsgLen
		tcpServer.MaxMsgLen = gate.MaxMsgLen
		tcpServer.LittleEndian = gate.LittleEndian
		tcpServer.NewAgent = func(conn *network.TCPConn) network.Agent {
			a := &agent{conn: conn, gate: gate}
			return a
		}
	}

	// if wsServer != nil {
	// 	wsServer.Start()
	// }
	if tcpServer != nil {
		tcpServer.Start()
	}
	<-closeSig
	// if wsServer != nil {
	// 	wsServer.Close()
	// }
	if tcpServer != nil {
		tcpServer.Close()
	}
}

func (gate *Gate) OnDestroy() {}

type agent struct {
	conn network.Conn
	gate *Gate
}

func (a *agent) Run() {
	for {
		data, err := a.conn.ReadMsg()
		if err != nil {
			log.Log.WithField("Err", err).Debug("read message")
			break
		}
		log.Log.WithField("data", len(data)).Debug()
		nsqproducer.NsqProducer.Publish("client", data)
	}
}

func (a *agent) OnClose() {

}

func (a *agent) WriteMsg(msg interface{}) {
	if a.gate.Processor != nil {
		data, err := a.gate.Processor.Marshal(msg)
		if err != nil {
			log.Log.WithFields(logrus.Fields{"MsgType": reflect.TypeOf(msg), "Err": err}).Error("marshal message")
			return
		}
		err = a.conn.WriteMsg(data...)
		if err != nil {
			log.Log.WithFields(logrus.Fields{"MsgType": reflect.TypeOf(msg), "Err": err}).Error("write message")
		}
	}
}

func (a *agent) LocalAddr() net.Addr {
	return a.conn.LocalAddr()
}

func (a *agent) RemoteAddr() net.Addr {
	return a.conn.RemoteAddr()
}

func (a *agent) Close() {
	a.conn.Close()
}

func (a *agent) Destroy() {
	a.conn.Destroy()
}
