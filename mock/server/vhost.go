package server

import "github.com/tiago4orion/amqputil"

// VHost is a fake AMQP virtual host
type VHost struct {
	name      string
	exchanges map[string]Exchange
	queues    map[string]Queue
}

type Channel VHost

// NewVHost create a new fake AMQP Virtual Host
func NewVHost(name string) *VHost {
	vh := VHost{
		name:   name,
		queues: make(map[string]Queue),
	}

	vh.createDefaultExchanges()
	return &vh
}

func (v *VHost) createDefaultExchanges() {
	exchs := make(map[string]Exchange)
	exchs["amqp.topic"] = &TopicExchange{}
	v.exchanges = exchs
}

func (v *VHost) Ack(tag uint64, multiple bool) error {
	return nil
}

func (v *VHost) Nack(tag uint64, multiple bool, requeue bool) error {
	return nil
}

func (v *VHost) Reject(tag uint64, requeue bool) error {
	return nil
}

func (v *VHost) Cancel(consumer string, noWait bool) error {
	return nil
}

func (v *VHost) ExchangeDeclare(name, kind string, opt amqputil.Option) error {
	return nil
}

func (v *VHost) QueueDeclare(name string, args amqputil.Option) (amqputil.Queue, error) {
	return &Queue{}, nil
}

func (v *VHost) QueueBind(name, key, exchange string, opt amqputil.Option) error {
	return nil
}

func (v *VHost) Consume(queue, consumer string, opt amqputil.Option) (<-chan amqputil.Delivery, error) {
	return nil, nil
}

func (v *VHost) Publish(exc, route string, msg []byte) error {
	return nil
}
