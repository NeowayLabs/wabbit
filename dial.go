package amqputil

import (
	"time"

	"github.com/streadway/amqp"
)

type AMQPConn struct {
	Conn        *amqp.Connection

	// closure info of connection
	dialFn      func() error
	attempts    uint8
}

func New() *AMQPConn {
	return &AMQPConn{}
}

func (conn *AMQPConn) Dial(uri string) error {
	conn.dialFn = func() error {
		var err error

		conn.Conn, err = amqp.Dial(uri)

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
func (conn *AMQPConn) AutoRedial(outChan chan error, onSuccess func()) {
	errChan := conn.Conn.NotifyClose(make(chan *amqp.Error))

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
				conn.attempts += 1
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
