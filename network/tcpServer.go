package network

import (
	"errors"
	"net"
	"sync"
	"time"

	"gateServer/log"

	"github.com/nsqio/go-nsq"
)

type TCPServer struct {
	Addr            string               // 监听网络地址
	MaxConnNum      int                  // 最大连接数
	PendingWriteNum int                  // 连接最大可写数
	NewAgent        func(*TCPConn) Agent // 代理
	ln              net.Listener
	conns           ConnSet // 连接集合
	mutexConns      sync.Mutex
	wgLn            sync.WaitGroup
	wgConns         sync.WaitGroup

	// msg parser
	LenMsgLen    int //
	MinMsgLen    uint32
	MaxMsgLen    uint32
	LittleEndian bool
	msgParser    *MsgParser

	// 连接代理池
	agentLen  uint32
	agentList []Agent

	// nsq消费者
	nsqConsumer *nsq.Consumer
}

func (server *TCPServer) Start() {
	server.init()
	go server.run()
}

func (server *TCPServer) HandleMessage(message *nsq.Message) error {
	return errors.New("sss")
}

func (server *TCPServer) init() {
	ln, err := net.Listen("tcp", server.Addr)
	if err != nil {
		log.Log.WithField("Error", err).Fatal("Listen Failed!")
	}

	if server.MaxConnNum <= 0 {
		server.MaxConnNum = 100
		log.Log.WithField("MaxConnNum", server.MaxConnNum).Info("Invalid MaxConnNum And Reset")
	}

	if server.PendingWriteNum <= 0 {
		server.PendingWriteNum = 100
		log.Log.WithField("PendingWriteNum", server.PendingWriteNum).Info("Invalid PendingWriteNum And Reset")
	}

	if server.NewAgent == nil {
		log.Log.Fatal("NewAgent must not be nil")
	}

	server.ln = ln
	server.conns = make(ConnSet)

	msgParser := NewMsgParser()
	msgParser.SetMsgLen(server.MinMsgLen, server.MaxMsgLen)
	msgParser.SetByteOrder(server.LittleEndian)
	server.msgParser = msgParser

	server.agentList = make([]Agent, server.MaxConnNum)

	nsqConfig := nsq.NewConfig()
	if server.nsqConsumer, err = nsq.NewConsumer("game", "main", nsqConfig); err != nil {
		log.Log.WithField("Error", err.Error()).Fatal("New Nsq Consumer Failed!")
	}

	server.nsqConsumer.AddHandler(server)
	if err = server.nsqConsumer.ConnectToNSQD("127.0.0.1:4150"); err != nil {
		log.Log.WithField("Error", err.Error()).Fatal("Nsq Connect Failed!")
	}
}

func (server *TCPServer) run() {
	server.wgLn.Add(1)
	defer server.wgLn.Done()

	var tempDelay time.Duration
	for {
		conn, err := server.ln.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				continue
			}
			return
		}
		tempDelay = 0

		server.mutexConns.Lock()
		if len(server.conns) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			log.Log.Debug("too many connections")
			continue
		}

		server.conns[conn] = struct{}{}
		server.mutexConns.Unlock()

		server.wgConns.Add(1)

		tcpConn := newTCPConn(conn, server.PendingWriteNum, server.msgParser)
		agent := server.NewAgent(tcpConn)
		go func() {
			agent.Run()

			tcpConn.Close()
			server.mutexConns.Lock()
			delete(server.conns, conn)
			server.mutexConns.Unlock()

			agent.OnClose()

			server.wgConns.Done()
		}()

		server.agentList[server.agentLen] = agent
		server.agentLen++
	}
}

func (server *TCPServer) Close() {
	server.ln.Close()
	server.wgLn.Wait()

	server.mutexConns.Lock()
	for conn := range server.conns {
		conn.Close()
	}

	server.conns = nil
	server.mutexConns.Unlock()
	server.wgConns.Wait()
}
