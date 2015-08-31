package rabbitmq

import "github.com/streadway/amqp"

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
