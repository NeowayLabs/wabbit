package rabbitmq

import "github.com/streadway/amqp"

type Delivery struct {
	*amqp.Delivery
}
