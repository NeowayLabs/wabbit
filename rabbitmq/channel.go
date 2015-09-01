package rabbitmq

import (
	"github.com/streadway/amqp"
	"github.com/tiago4orion/amqputil"
)

type Channel struct {
	*amqp.Channel
}

func (ch *Channel) Publish(exc, route string, msg []byte) error {
	return ch.Channel.Publish(
		exc,   // publish to an exchange
		route, // routing to 0 or more queues
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(msg),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	)
}

func (ch *Channel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args interface{}) (<-chan amqputil.Delivery, error) {
	amqpd, err := ch.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args.(amqp.Table))

	if err != nil {
		return nil, err
	}

	deliveries := make(chan amqputil.Delivery)

	go func() {
		for d := range amqpd {
			deliveries <- Delivery{&d}
		}
	}()

	return deliveries, nil
}

func (ch *Channel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args interface{}) error {
	return ch.Channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args.(amqp.Table))
}

func (ch *Channel) QueueBind(name, key, exchange string, noWait bool, args interface{}) error {
	return ch.Channel.QueueBind(name, key, exchange, noWait, args.(amqp.Table))
}

func (ch *Channel) QueueDeclare(name string, durable, autoDelete, exclusive, noWait bool, args interface{}) (amqputil.Queue, error) {
	q, err := ch.Channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args.(amqp.Table))

	if err != nil {
		return nil, err
	}

	return &Queue{&q}, nil
}