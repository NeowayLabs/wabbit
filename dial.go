package amqputil

import (
	"errors"
	"time"

	"github.com/streadway/amqp"
)

type AMQPConn struct {
	Conn     *amqp.Connection
	dialFn   func() error
	attempts uint8
	IsConnected bool
}

func New() *AMQPConn {
	return &AMQPConn{}
}

func (conn *AMQPConn) Dial(uri string) error {
	conn.dialFn = func() error {
		var err error

		conn.IsConnected = false
		conn.Conn, err = amqp.Dial(uri)

		if err != nil {
			return err
		}

		conn.IsConnected = true
		return nil
	}

	return conn.dialFn()
}

func (conn *AMQPConn) AutoRedial(outChan chan error) {
	errChan := conn.Conn.NotifyClose(make(chan *amqp.Error))

	go func() {
		var err error

		select {
		case amqpErr := <-errChan:
			err = amqpErr
		attempt:
			if err != nil {
				conn.IsConnected = false
				outChan <- errors.New(err.Error())
			}

			if conn.attempts > 60 {
				conn.attempts = 0
			}

			// Wait n Seconds where n == conn.attempts...
			time.Sleep(time.Duration(int64(conn.attempts) * int64(time.Second)))

			err = conn.dialFn()

			if err != nil {
				conn.attempts += 1
				goto attempt
			}

			// enabled AutoRedial on the new connection
			conn.AutoRedial(outChan)
		}
	}()
}
