package mock

import (
	"sync"
	"time"

	"github.com/tiago4orion/amqputil"
)

const (
	// 1 second
	defaultReconnectDelay = 1
)

// Conn is the fake AMQP connection
type Conn struct {
	isConnected   bool
	errChan       chan error
	mu            *sync.Mutex
	hasAutoRedial bool
	amqpServer    *AMQPServer

	dialFn func() error
}

// NewConn creates a new fake connection
func NewConn() *Conn {
	return &Conn{
		errChan: make(chan error),
		mu:      &sync.Mutex{},
	}
}

// Dial mock the connection dialing to rabbitmq
func (conn *Conn) Dial(amqpuri string) error {
	conn.dialFn = func() error {
		var err error

		conn.amqpServer, err = connect(amqpuri)

		if err != nil {
			return err
		}

		// concurrent access with Close method
		conn.mu.Lock()
		conn.isConnected = true
		conn.mu.Unlock()
		return nil
	}

	return conn.dialFn()
}

// AutoRedial mock the reconnection faking a delay of 1 second
func (conn *Conn) AutoRedial(outChan chan error, done chan bool) {
	conn.mu.Lock()
	conn.hasAutoRedial = true
	conn.mu.Unlock()

	go func(errChan <-chan error) {
		var err error
		var attempts uint

		select {
		case amqpErr := <-errChan:
			err = amqpErr

			if amqpErr == nil {
				// Gracefull connection close
				return
			}
		lattempts:
			outChan <- err

			if attempts > 60 {
				attempts = 0
			}

			// Wait n Seconds where n == attempts...
			time.Sleep(time.Duration(attempts) * time.Second)

			err = conn.dialFn()

			if err != nil {
				attempts++
				goto lattempts
			}

			attempts = 0

			// enabled AutoRedial on the new connection
			conn.AutoRedial(outChan, done)
			done <- true
			return
		}
	}(conn.errChan)
}

// Close the fake connection
func (conn *Conn) Close() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	conn.isConnected = false
	conn.amqpServer = nil
	// enables AutoRedial to gracefully shutdown
	// This isn't amqputil stuff. It's the streadway/amqp way of notify the shutdown
	if conn.hasAutoRedial {
		conn.errChan <- nil
	}

	return nil
}

// Channel creates a new fake channel
func (conn *Conn) Channel() (amqputil.Channel, error) {
	return &Channel{}, nil
}
