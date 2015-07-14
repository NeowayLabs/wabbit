package amqputil

import (
	"time"

	"github.com/streadway/amqp"
)

// Conn is the amqp connection
type Conn struct {
	*amqp.Connection

	// closure info of connection
	dialFn   func() error
	attempts uint8
}

// New do what the fucking name says...
func New() *Conn {
	return &Conn{}
}

// Dial to rabbitmq
func (conn *Conn) Dial(uri string) error {
	// closure the uri for handle reconnects
	conn.dialFn = func() error {
		var err error

		conn.Connection, err = amqp.Dial(uri)

		if err != nil {
			return err
		}

		return nil
	}

	return conn.dialFn()
}

// AutoRedial manages the automatic redial of connection when unexpected closed.
// outChan is an unbuffered channel required to receive the errors that results from
// attempts of reconnect. On successfully reconnected, the function onSuccess
// is invoked.
//
// The outChan parameter can receive *amqp.Error for AMQP connection errors
// or errors.Error for any other net/tcp internal error.
//
// Redial strategy:
// If the connection is closed in an unexpected way (opposite of conn.Close()), then
// AutoRedial will try to automatically reconnect waiting for N seconds before each
// attempt, where N is the number of attempts of reconnecting. If the number of
// attempts reach 60, it will be zero'ed.
func (conn *Conn) AutoRedial(outChan chan error, onSuccess func()) {
	errChan := conn.NotifyClose(make(chan *amqp.Error))

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
