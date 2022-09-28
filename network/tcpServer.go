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
	Addr            string // 监听网络地址
	MaxConnNum      int    // 最大连接数
	PendingWriteNum int    // 连接最大可写数
	ln              net.Listener
	mutexConns      sync.Mutex
	wgLn            sync.WaitGroup
	wgConns         sync.WaitGroup

	// msg parser
	LenMsgLen    int //
	MinMsgLen    uint32
	MaxMsgLen    uint32
	LittleEndian bool

	// 连接代理池
	AgentMap map[int64]*agent
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

	server.ln = ln

	server.AgentMap = make(map[int64]*agent)
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

		if int(Connections.count) >= server.MaxConnNum {
			server.mutexConns.Unlock()
			conn.Close()
			log.Log.Debug("too many connections")
			continue
		}

		server.wgConns.Add(1)

		agent := newAgent(conn)
		go func() {
			agent.Run()

			agent.OnClose()

			server.wgConns.Done()
		}()
		server.AgentMap[agent.ID] = agent
	}
}

func (server *TCPServer) GetAgent(agentID int64) *agent {
	agent, ok := server.AgentMap[agentID]
	if !ok {
		return nil
	}

	return agent
}

func (server *TCPServer) Close() {
	server.ln.Close()
	server.wgLn.Wait()

	server.mutexConns.Lock()
	for _, agent := range server.AgentMap {
		agent.OnClose()
	}

	server.AgentMap = nil
	server.mutexConns.Unlock()
	server.wgConns.Wait()
}
