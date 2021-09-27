package server

import (
	"fmt"
	"sync"

	"github.com/NeowayLabs/wabbit"
)

// VHost is a fake AMQP virtual host
type VHost struct {
	name      string
	mu        sync.Mutex // Protects exchanges and queues.
	exchanges map[string]Exchange
	queues    map[string]*Queue
}

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
	exchs["amq.topic"] = NewTopicExchange("amq.topic")
	exchs["amq.direct"] = NewDirectExchange("amq.direct")
	exchs["topic"] = NewTopicExchange("topic")
	exchs["direct"] = NewDirectExchange("direct")
	exchs[""] = NewDirectExchange("amq.direct")

	v.mu.Lock()
	v.exchanges = exchs
	v.mu.Unlock()
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
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.exchangeDeclare(name, kind, false, opt)
}

func (v *VHost) ExchangeDeclarePassive(name, kind string, opt wabbit.Option) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.exchangeDeclare(name, kind, true, opt)
}

func (v *VHost) exchangeDeclare(name, kind string, passive bool, opt wabbit.Option) error {
	if _, ok := v.exchanges[name]; ok {
		// TODO: We need review this. If the application is trying to re-create an exchange
		// using other options we shall not return NIL because this indicates success,
		// but we didn't declared anything.
		// The AMQP 0.9.1 spec says nothing about that. It only says that AMQP uses the
		// "declare" concept instead of the "create" concept. If something is already
		// declared it's no problem...
		return nil
	}

	if passive {
		return fmt.Errorf("exception (404) Reason: \"NOT_FOUND - no exchange '%s' in vhost '%s'\"", name, v.name)
	}

	switch kind {
	case "topic":
		v.exchanges[name] = NewTopicExchange(name)
	case "direct":
		v.exchanges[name] = NewDirectExchange(name)
	case "headers":
		v.exchanges[name] = NewHeadersExchange(name)
	default:
		return fmt.Errorf("invalid exchange type: %s", kind)
	}

	return nil
}

func (v *VHost) QueueDeclare(name string, args wabbit.Option) (wabbit.Queue, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.queueDeclare(name, false, args)
}

func (v *VHost) QueueDeclarePassive(name string, args wabbit.Option) (wabbit.Queue, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.queueDeclare(name, true, args)
}

func (v *VHost) queueDeclare(name string, passive bool, args wabbit.Option) (wabbit.Queue, error) {
	if q, ok := v.queues[name]; ok {
		return q, nil
	}

	if passive {
		return nil, fmt.Errorf("exception (404) Reason: \"NOT_FOUND - no queue '%s' in vhost '%s'\"", name, v.name)
	}

	q := NewQueue(name)

	v.queues[name] = q

	err := v.queueBind(name, name, "", nil)

	if err != nil {
		return nil, err
	}

	return q, nil
}

func (v *VHost) QueueDelete(name string, args wabbit.Option) (int, error) {
	v.mu.Lock()
	delete(v.queues, name)
	v.mu.Unlock()

	return 0, nil
}

func (v *VHost) QueueBind(name, key, exchange string, options wabbit.Option) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.queueBind(name, key, exchange, options)
}

func (v *VHost) queueBind(name, key, exchange string, _ wabbit.Option) error {
	var (
		exch Exchange
		q    *Queue
		ok   bool
	)

	if exch, ok = v.exchanges[exchange]; !ok {
		return fmt.Errorf("unknown exchange '%s'", exchange)
	}

	if q, ok = v.queues[name]; !ok {
		return fmt.Errorf("unknown queue '%s'", name)
	}

	exch.addBinding(key, &BindingsMap{q, nil})
	return nil
}

func (v *VHost) QueueUnbind(name, key, exchange string, options wabbit.Option) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.queueUnbind(name, key, exchange, options)
}

func (v *VHost) queueUnbind(name, key, exchange string, _ wabbit.Option) error {
	var (
		exch Exchange
		ok   bool
	)

	if exch, ok = v.exchanges[exchange]; !ok {
		return fmt.Errorf("unknown exchange '%s'", exchange)
	}

	if _, ok = v.queues[name]; !ok {
		return fmt.Errorf("unknown queue '%s'", name)
	}

	exch.delBinding(key)
	return nil
}

// Publish push a new message to queue data channel.
// The queue data channel is a buffered channel of length `QueueMaxLen`. If
// the queue is full, this method will block until some messages are consumed.
func (v *VHost) Publish(exc, route string, d *Delivery, options wabbit.Option) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.publish(exc, route, d, options)
}

func (v *VHost) publish(exc, route string, d *Delivery, _ wabbit.Option) error {
	var (
		exch Exchange
		ok   bool
		err  error
	)

	if exch, ok = v.exchanges[exc]; !ok {
		return fmt.Errorf("unknow exchange '%s'", exc)
	}

	err = exch.route(route, d)

	if err != nil {
		return err
	}

	return nil
}
