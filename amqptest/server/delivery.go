package server

type (
	// Delivery is an interface to delivered messages
	Delivery struct {
		data          []byte
		tag           uint64
		consumerTag   string
		originalRoute string
		channel       *Channel
	}
)

func NewDelivery(ch *Channel, data []byte, tag uint64) *Delivery {
	return &Delivery{
		data:    data,
		channel: ch,
		tag:     tag,
	}
}

func (d *Delivery) Ack(multiple bool) error {
	return d.channel.Ack(d.tag, multiple)
}

func (d *Delivery) Nack(multiple, requeue bool) error {
	return d.channel.Nack(d.tag, multiple, requeue)
}

func (d *Delivery) Reject(requeue bool) error {
	return d.channel.Nack(d.tag, false, requeue)
}

func (d *Delivery) Body() []byte {
	return d.data
}

func (d *Delivery) DeliveryTag() uint64 {
	return d.tag
}

func (d *Delivery) ConsumerTag() string {
	return d.consumerTag
}
