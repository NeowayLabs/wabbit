package client

import (
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/tiago4orion/wabbit"
	"github.com/tiago4orion/wabbit/mock/server"
)

const (
	// 1 second
	defaultReconnectDelay = 1
)

// Conn is the fake AMQP connection
type Conn struct {
	isConnected   bool
	connID        string
	errChan       chan error
	defErrDone    chan bool
	mu            *sync.Mutex
	hasAutoRedial bool
	amqpServer    *server.AMQPServer

	dialFn func() error
}

// NewConn creates a new fake connection
func NewConn() *Conn {
	return &Conn{
		errChan:    make(chan error),
		defErrDone: make(chan bool),
		mu:         &sync.Mutex{},
	}
}

// Dial mock the connection dialing to rabbitmq
func (conn *Conn) Dial(amqpuri string) error {
	conn.dialFn = func() error {
		var err error
		conn.connID = uuid.New()
		conn.amqpServer, err = server.Connect(amqpuri, conn.connID, conn.errChan)

		if err != nil {
			return err
		}

		// concurrent access with Close method
		conn.mu.Lock()
		conn.isConnected = true
		conn.mu.Unlock()

		// by default, we discard any errors
		// send something to defErrDone to destroy
		// this goroutine and start consume the errors
		go func() {
			for {
				select {
				case <-conn.errChan:
				case <-conn.defErrDone:
					return
				}
			}
		}()

		return nil
	}

	return conn.dialFn()
}

// AutoRedial mock the reconnection faking a delay of 1 second
func (conn *Conn) AutoRedial(outChan chan error, done chan bool) {
	if !conn.hasAutoRedial {
		conn.defErrDone <- true
		conn.mu.Lock()
		conn.hasAutoRedial = true
		conn.mu.Unlock()
	}

	go func() {
		var err error
		var attempts uint

		select {
		case amqpErr := <-conn.errChan:
			err = amqpErr

			if amqpErr == nil {
				// Gracefull connection close
				return
			}
		lattempts:
			// send the error to client
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
	}()
}

// Close the fake connection
func (conn *Conn) Close() error {
	conn.mu.Lock()
	defer conn.mu.Unlock()

	conn.isConnected = false
	conn.amqpServer = nil

	// enables AutoRedial to gracefully shutdown
	// This isn't wabbit stuff. It's the streadway/amqp way of notify the shutdown
	if conn.hasAutoRedial {
		conn.errChan <- nil
	} else {
		conn.defErrDone <- true
	}

	return nil
}

// Channel creates a new fake channel
func (conn *Conn) Channel() (wabbit.Channel, error) {
	return conn.amqpServer.CreateChannel()
}
