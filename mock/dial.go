package mock

import (
	"errors"
	"sync"
	"time"

	"github.com/tiago4orion/amqputil"
)

const (
	// 1 second
	defaultReconnectDelay = 1
)

var (
	defConn *Conn
)

func init() {
	defConn = newConn()
}

// Conn is the fake AMQP connection
type Conn struct {
	running       bool
	isConnected   bool
	errChan       chan error
	mu            *sync.Mutex
	hasAutoRedial bool

	dialFn func() error
}

func newConn() *Conn {
	return &Conn{
		errChan: make(chan error),
		mu:      &sync.Mutex{},
	}
}

// New creates the fake conn
func New() *Conn {
	return defConn
}

// Dial mock the connection dialing to rabbitmq
func (conn *Conn) Dial(amqpuri string) error {
	conn.dialFn = func() error {
		conn.mu.Lock()
		defer conn.mu.Unlock()

		if conn.running {
			conn.isConnected = true
			return nil
		}

		return errors.New("Failed to connect to '" + amqpuri + "'")
	}

	return conn.dialFn()
}

// AutoRedial mock the reconnection faking a delay of 1 second
func (conn *Conn) AutoRedial(outChan chan error, onSuccess func()) {
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
			conn.AutoRedial(outChan, onSuccess)
			onSuccess()
			return
		}
	}(conn.errChan)
}

// Close the fake connection
func (conn *Conn) Close() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	conn.isConnected = false
	// enables AutoRedial to gracefully shutdown
	// This isn't amqputil stuff. It's the streadway/amqp way of notify the shutdown
	conn.errChan <- nil
	return nil
}

// Channel creates a new fake channel
func (conn *Conn) Channel() (amqputil.Channel, error) {
	return &Channel{}, nil
}

// StopServer stops the fake server
func (conn *Conn) StopServer() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.running = false
}

// StartServer stops the fake server
func (conn *Conn) StartServer() {
	conn.mu.Lock()
	defer conn.mu.Unlock()
	conn.running = true

	// avoid deadlock
	if conn.hasAutoRedial {
		conn.errChan <- errors.New("rabbitmq disconnected")
	}
}

// StartRabbitmq starts the fake AMQP server
func StartServer() {
	defConn.StartServer()
}

// StopRabbitmq stop the fake AMQP server
func StopServer() {
	defConn.StopServer()
}
