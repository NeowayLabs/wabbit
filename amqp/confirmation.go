package amqp

import amqp "github.com/rabbitmq/amqp091-go"

type Confirmation struct {
	amqp.Confirmation
}

func (c Confirmation) Ack() bool {
	return c.Confirmation.Ack
}

func (c Confirmation) DeliveryTag() uint64 {
	return c.Confirmation.DeliveryTag
}
