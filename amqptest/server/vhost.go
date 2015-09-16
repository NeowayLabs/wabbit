package server

import (
	"fmt"
	"os"
	"sync/atomic"

	"github.com/tiago4orion/wabbit"
)

var consumerSeq uint64

func uniqueConsumerTag() string {
	return fmt.Sprintf("ctag-%s-%d", os.Args[0], atomic.AddUint64(&consumerSeq, 1))
}

// VHost is a fake AMQP virtual host
type VHost struct {
	name      string
	exchanges map[string]Exchange
	queues    map[string]*Queue
	consumers map[string]consumer
}

type Channel VHost

type consumer struct {
	tag        string
	deliveries chan wabbit.Delivery
	done       chan bool
}

// NewVHost create a new fake AMQP Virtual Host
func NewVHost(name string) *VHost {
	vh := VHost{
		name:      name,
		queues:    make(map[string]*Queue),
		exchanges: make(map[string]Exchange),
		consumers: make(map[string]consumer),
	}

	vh.createDefaultExchanges()
	return &vh
}

func (v *VHost) createDefaultExchanges() {
	exchs := make(map[string]Exchange)
	exchs["amq.topic"] = &TopicExchange{}
	exchs["amq.direct"] = &DirectExchange{}
	exchs["topic"] = &TopicExchange{}
	exchs["direct"] = &DirectExchange{}
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

// Qos isn't implemented in the fake server
func (v *VHost) Qos(prefetchCount, prefetchSize int, global bool) error {
	// do nothing. It's a implementation-specific tuning
	return nil
}

func (v *VHost) ExchangeDeclare(name, kind string, opt wabbit.Option) error {
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
	case "topic":
		v.exchanges[name] = NewTopicExchange(name)
	case "direct":
		v.exchanges[name] = NewDirectExchange(name)
	default:
		return fmt.Errorf("Invalid exchange type: %s", kind)
	}

	return nil
}

func (v *VHost) QueueDeclare(name string, args wabbit.Option) (wabbit.Queue, error) {
	q := NewQueue(name)

	v.queues[name] = q
	return q, nil
}

func (v *VHost) QueueBind(name, key, exchange string, _ wabbit.Option) error {
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

// Consume starts a fake consumer of queue
func (v *VHost) Consume(queue, consumerName string, _ wabbit.Option) (<-chan wabbit.Delivery, error) {
	var (
		found bool
		c     consumer
	)

	if consumerName == "" {
		consumerName = uniqueConsumerTag()
	}

	if c, found = v.consumers[consumerName]; found {
		c.done <- true
		close(c.deliveries)
	}

	c = consumer{
		tag:        consumerName,
		deliveries: make(chan wabbit.Delivery),
	}

	v.consumers[consumerName] = c

	q, ok := v.queues[queue]

	if !ok {
		return nil, fmt.Errorf("Unknown queue '%s'", queue)
	}

	go func() {
		for {
			select {
			case <-c.done:
				close(c.deliveries)
				return
			case d := <-q.data:
				c.deliveries <- d
			}
		}
	}()

	return c.deliveries, nil
}

// Publish push a new message to queue data channel.
// The queue data channel is a buffered channel of length `QueueMaxLen`. If
// the queue is full, this method will block until some messages are consumed.
func (v *VHost) Publish(exc, route string, msg []byte, _ wabbit.Option) error {
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
