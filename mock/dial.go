package mock

import (
	"sync"
	"time"
)

const (
	// 1 second
	defaultReconnectDelay = 1
)

type Conn struct {
	isConnected bool
	l           *sync.Mutex

	redialFn func() error
}

func New() *Conn {
	return &Conn{}
}

// Dial mock the connection dialing to rabbitmq
func (conn *Conn) Dial(amqpuri string) error {
	conn.redialFn = func() error {
		conn.l.Lock()
		defer conn.l.Unlock()

		conn.isConnected = true
		return nil
	}

	return conn.redialFn()
}

// AutoRedial mock the reconnection faking a delay of 1 second
func (conn *Conn) AutoRedial(outChan chan error, success func()) {
	time.Sleep(time.Second * defaultReconnectDelay)
	conn.redialFn()
}

func (conn *Conn) Channel() *Channel {
	return &Channel{}
}
