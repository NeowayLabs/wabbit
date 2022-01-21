package amqp

import (
	"time"

	"github.com/NeowayLabs/wabbit"
	amqp "github.com/rabbitmq/amqp091-go"
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

func (d *Delivery) MessageId() string {
	return d.Delivery.MessageId
}

func (d *Delivery) Timestamp() time.Time {
	return time.Now()
}

func (d *Delivery) ContentType() string {
	return d.Delivery.ContentType
}
