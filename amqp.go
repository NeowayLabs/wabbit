// Package amqputil provides some abstractions for AMQP for easy the testing.
// The best way to test
package amqputil

type (
	Conn interface{
		Channel() (Channel, error)
		Dial(amqpuri string) error
		AutoRedial(errChan chan error, cbk func())
	}

	Channel interface{
		Ack(tag uint64, multiple bool) error
		Nack(tag uint64, multiple bool, requeue bool) error
		Reject(tag uint64, requeue bool) error
		
		Publisher
	}

	Publisher interface{
		Publish(exc, route string, msg []byte) error
	}
)

