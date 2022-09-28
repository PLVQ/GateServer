package network

import (
	"sync/atomic"
)

// Connections is a global variable which is used by session.
var Connections = newConnectionService()

type connectionService struct {
	count int64
	sid   int64
}

func newConnectionService() *connectionService {
	return &connectionService{sid: 0}
}

// Increment increment the connection count
func (c *connectionService) Increment() {
	atomic.AddInt64(&c.count, 1)
}

// Decrement decrement the connection count
func (c *connectionService) Decrement() {
	atomic.AddInt64(&c.count, -1)
}

// Count returns the connection numbers in current
func (c *connectionService) Count() int64 {
	return atomic.LoadInt64(&c.count)
}

// Reset reset the connection service status
func (c *connectionService) Reset() {
	atomic.StoreInt64(&c.count, 0)
	atomic.StoreInt64(&c.sid, 0)
}

// SessionID returns the session id
func (c *connectionService) SessionID() int64 {
	return atomic.AddInt64(&c.sid, 1)
}
