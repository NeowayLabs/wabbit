package amqp

import (
	"github.com/NeowayLabs/wabbit"
	"github.com/streadway/amqp"
)

type Delivery struct {
	*amqp.Delivery
}

func (d *Delivery) Body() []byte {
	return d.Delivery.Body
}

func (d *Delivery) Headers() wabbit.Option {
	return wabbit.Option(d.Delivery.Headers)
}

func (d *Delivery) DeliveryTag() uint64 {
	return d.Delivery.DeliveryTag
}

func (d *Delivery) ConsumerTag() string {
	return d.Delivery.ConsumerTag
}
