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
	d.channel.Ack(d.tag, multiple)
	return nil
}

func (d *Delivery) Nack(multiple, request bool) error {
	return nil
}

func (d *Delivery) Reject(requeue bool) error {
	return nil
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
