package server

import (
	"fmt"

	"github.com/tiago4orion/amqputil"
)

// VHost is a fake AMQP virtual host
type VHost struct {
	name      string
	exchanges map[string]Exchange
	queues    map[string]*Queue
}

type Channel VHost

// NewVHost create a new fake AMQP Virtual Host
func NewVHost(name string) *VHost {
	vh := VHost{
		name:      name,
		queues:    make(map[string]*Queue),
		exchanges: make(map[string]Exchange),
	}

	vh.createDefaultExchanges()
	return &vh
}

func (v *VHost) createDefaultExchanges() {
	exchs := make(map[string]Exchange)
	exchs["amq.topic"] = &TopicExchange{}
	exchs["amq.direct"] = &DirectExchange{}
	exchs[""] = &DirectExchange{
		name: "amq.direct",
	}

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
	if _, ok := v.exchanges[name]; ok {
		// TODO: We need review this. If the application is trying to re-create an exchange
		// using other options we shall not return NIL because this indicates success,
		// but we didn't declared anything.
		// The AMQP 0.9.1 spec says nothing about that. It only says that AMQP uses the
		// "declare" concept instead of the "create" concept. If something is already
		// declared it's no problem...
		return nil
	}

	switch kind {
	case "amq.topic":
		v.exchanges[name] = NewTopicExchange(name)
	case "amq.direct":
		v.exchanges[name] = NewDirectExchange(name)
	default:
		return fmt.Errorf("Invalid exchange type: %s", kind)
	}

	return nil
}

func (v *VHost) QueueDeclare(name string, args amqputil.Option) (amqputil.Queue, error) {
	q := &Queue{
		name: name,
	}

	v.queues[name] = q
	return q, nil
}

func (v *VHost) QueueBind(name, key, exchange string, _ amqputil.Option) error {
	var (
		exch Exchange
		q    *Queue
		ok   bool
	)

	if exch, ok = v.exchanges[exchange]; !ok {
		return fmt.Errorf("Unknown exchange '%s'", exchange)
	}

	if q, ok = v.queues[name]; !ok {
		return fmt.Errorf("Unknown queue '%s'", name)
	}

	exch.addBinding(key, q)
	return nil
}

func (v *VHost) Consume(queue, consumer string, opt amqputil.Option) (<-chan amqputil.Delivery, error) {
	q, ok := v.queues[queue]

	if !ok {
		return nil, fmt.Errorf("Unknown queue '%s'", queue)
	}

	return q.data, nil
}

func (v *VHost) Publish(exc, route string, msg []byte, _ amqputil.Option) error {
	var (
		exch Exchange
		ok   bool
		q    *Queue
		err  error
	)

	if exch, ok = v.exchanges[exc]; !ok {
		return fmt.Errorf("Unknow exchange '%s'", exc)
	}

	q, err = exch.route(route, msg)

	if err != nil {
		return err
	}

	q.data <- NewDelivery(msg)
	return nil
}
