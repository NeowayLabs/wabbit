package mock

import "github.com/tiago4orion/amqputil"

type (
	Channel struct{}
)

func (ch *Channel) Ack(tag uint64, multiple bool) error {
	return nil
}

func (ch *Channel) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func (ch *Channel) Reject(tag uint64, requeue bool) error {
	return nil
}

func (ch *Channel) Cancel(consumer string, noWait bool) error {
	return nil
}

func (ch *Channel) Publish(exc, route string, msg []byte) error {
	return nil
}

func (ch *Channel) Consume(queue, consumer string, opt amqputil.Option) (<-chan amqputil.Delivery, error) {
	return nil, nil
}

func (ch *Channel) ExchangeDeclare(name, kind string, opt amqputil.Option) error {
	return nil
}

func (ch *Channel) QueueDeclare(name string, args amqputil.Option) (amqputil.Queue, error) {
	return NewQueue(name), nil
}

func (ch *Channel) QueueBind(name, key, exchange string, opt amqputil.Option) error {
	return nil
}
