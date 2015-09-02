// Package amqputil provides some abstractions for AMQP for easy the testing.
// The best way to test
package amqputil

type (
	Option map[string]interface{}

	Conn interface{
		Channel() (Channel, error)
		Dial(amqpuri string) error
		AutoRedial(errChan chan error, cbk func())
		Close() error
	}

	Channel interface{
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

	Queue interface{
		Name() string
		Messages() int
		Consumers() int
	}

	Publisher interface{
		Publish(exc, route string, msg []byte) error
	}

	Delivery interface {
		Ack(multiple bool) error
		Nack(multiple, request bool) error
		Reject(requeue bool) error

		Body() []byte
		DeliveryTag() uint64
	}
)

