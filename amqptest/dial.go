package amqptest

import (
	"sync"
	"time"

	"code.google.com/p/go-uuid/uuid"

	"github.com/tiago4orion/wabbit"
	"github.com/tiago4orion/wabbit/amqptest/server"
	"github.com/tiago4orion/wabbit/utils"
)

const (
	// 1 second
	defaultReconnectDelay = 1
)

// Conn is the fake AMQP connection
type Conn struct {
	isConnected   bool
	connID        string
	errSpread     *utils.ErrBroadcast
	errChan       chan wabbit.Error
	defErrDone    chan bool
	mu            *sync.Mutex
	hasAutoRedial bool
	amqpServer    *server.AMQPServer

	dialFn func() error
}

// Dial mock the connection dialing to rabbitmq and
// returns the established connection or error if something goes wrong
func Dial(amqpuri string) (*Conn, error) {
	conn := &Conn{
		errSpread:  utils.NewErrBroadcast(),
		errChan:    make(chan wabbit.Error),
		defErrDone: make(chan bool),
		mu:         &sync.Mutex{},
	}

	conn.errSpread.Add(conn.errChan)

	conn.dialFn = func() error {
		var err error
		conn.connID = uuid.New()
		conn.amqpServer, err = server.Connect(amqpuri, conn.connID, conn.errSpread)

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

	err := conn.dialFn()

	if err != nil {
		return nil, err
	}

	return conn, nil
}

// NotifyClose publishs notifications about server or client errors in the given channel
func (conn *Conn) NotifyClose(c chan wabbit.Error) chan wabbit.Error {
	conn.errSpread.Add(c)
	return c
}

// AutoRedial mock the reconnection faking a delay of 1 second
func (conn *Conn) AutoRedial(outChan chan wabbit.Error, done chan bool) {
	if !conn.hasAutoRedial {
		conn.defErrDone <- true
		conn.mu.Lock()
		conn.hasAutoRedial = true
		conn.mu.Unlock()
	}

	go func() {
		var err wabbit.Error
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

			connErr := conn.dialFn()

			if connErr != nil {
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
		conn.errSpread.Write(nil)
	} else {
		conn.defErrDone <- true
	}

	return nil
}

// Channel creates a new fake channel
func (conn *Conn) Channel() (wabbit.Channel, error) {
	return conn.amqpServer.CreateChannel(conn.connID)
}
