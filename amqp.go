// Package amqputil provides some abstractions for AMQP for easy the testing.
// The best way to test
package amqputil

type (
	// Option is a map of AMQP configurations
	Option map[string]interface{}

	Conn interface {
		Channel() (Channel, error)
		Dial(amqpuri string) error
		AutoRedial(errChan chan error, cbk func())
		Close() error
	}

	// Channel is a AMQP channel interface
	Channel interface {
		Ack(tag uint64, multiple bool) error
		Nack(tag uint64, multiple bool, requeue bool) error
		Reject(tag uint64, requeue bool) error

		Cancel(consumer string, noWait bool) error
		ExchangeDeclare(name, kind string, opt Option) error
		QueueDeclare(name string, args Option) (Queue, error)
		QueueBind(name, key, exchange string, opt Option) error
		Consume(queue, consumer string, opt Option) (<-chan Delivery, error)
		Publisher
	}

	// Queue is a AMQP queue interface
	Queue interface {
		Name() string
		Messages() int
		Consumers() int
	}

	// Publisher is an interface to something able to publish messages
	Publisher interface {
		Publish(exc, route string, msg []byte) error
	}

	// Delivery is an interface to delivered messages
	Delivery interface {
		Ack(multiple bool) error
		Nack(multiple, request bool) error
		Reject(requeue bool) error

		Body() []byte
		DeliveryTag() uint64
	}
)
