package gate

import (
	"time"

	"gateServer/network"
	"gateServer/nsqer"
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
	}

	go nsqer.InitConsuemr("game", "game", &ClientMsgConsumer{Server: tcpServer})
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
