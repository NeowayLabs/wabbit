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
	// the fake rabbitmq server is running, isn't ?
	Running bool
	errChan chan error
	mu      *sync.Mutex
)

func init() {
	errChan = make(chan error)
	mu = &sync.Mutex{}
}

type Conn struct {
	isConnected bool
	attempts    uint

	dialFn func() error
}

func New() *Conn {
	return &Conn{}
}

// Dial mock the connection dialing to rabbitmq
func (conn *Conn) Dial(amqpuri string) error {
	conn.dialFn = func() error {
		mu.Lock()
		defer mu.Unlock()

		if Running {
			conn.isConnected = true
			return nil
		}

		return errors.New("Failed to connect to '" + amqpuri + "'")
	}

	return conn.dialFn()
}

// AutoRedial mock the reconnection faking a delay of 1 second
func (conn *Conn) AutoRedial(outChan chan error, onSuccess func()) {
	go func() {
		var err error

		select {
		case amqpErr := <-errChan:
			err = amqpErr

			if amqpErr == nil {
				// Gracefull connection close
				return
			}
		attempts:
			outChan <- err

			if conn.attempts > 60 {
				conn.attempts = 0
			}

			// Wait n Seconds where n == conn.attempts...
			time.Sleep(time.Duration(conn.attempts) * time.Second)

			err = conn.dialFn()

			if err != nil {
				conn.attempts++
				goto attempts
			}

			conn.attempts = 0

			// enabled AutoRedial on the new connection
			conn.AutoRedial(outChan, onSuccess)
			onSuccess()
			return
		}
	}()
}

func (conn *Conn) Close() error {
	conn.isConnected = false
	// enables AutoRedial to gracefully shutdown
	// This isn't amqputil stuff. It's the streadway/amqp way of notify the shutdown
	errChan <- nil
	return nil
}

func (conn *Conn) Channel() (amqputil.Channel, error) {
	return &Channel{}, nil
}

func StartRabbitmq() {
	mu.Lock()

	defer mu.Unlock()
	Running = true
}

func StopRabbitmq() {
	mu.Lock()
	Running = false
	mu.Unlock()

	go func() {
		errChan <- errors.New("rabbitmq disconnected")
	}()
}
