package rabbitmq

import (
	"errors"

	"github.com/streadway/amqp"
	"github.com/tiago4orion/amqputil"
)

type Channel struct {
	*amqp.Channel
}

func (ch *Channel) Publish(exc, route string, msg []byte) error {
	return ch.Channel.Publish(
		exc,   // publish to an exchange
		route, // routing to 0 or more queues
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(msg),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	)
}

func (ch *Channel) Consume(queue, consumer string, opt amqputil.Option) (<-chan amqputil.Delivery, error) {
	var (
		autoAck, exclusive, noLocal, noWait bool
		args                                amqp.Table
	)

	if v, ok := opt["autoAck"]; ok {
		autoAck, ok = v.(bool)

		if !ok {
			return nil, errors.New("durable option is of type bool")
		}
	}

	if v, ok := opt["exclusive"]; ok {
		exclusive, ok = v.(bool)

		if !ok {
			return nil, errors.New("exclusive option is of type bool")
		}
	}

	if v, ok := opt["noLocal"]; ok {
		noLocal, ok = v.(bool)

		if !ok {
			return nil, errors.New("noLocal option is of type bool")
		}
	}

	if v, ok := opt["noWait"]; ok {
		noWait, ok = v.(bool)

		if !ok {
			return nil, errors.New("noWait option is of type bool")
		}
	}

	if v, ok := opt["args"]; ok {
		args, ok = v.(amqp.Table)

		if !ok {
			return nil, errors.New("args is of type amqp.Table")
		}
	}

	amqpd, err := ch.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)

	if err != nil {
		return nil, err
	}

	deliveries := make(chan amqputil.Delivery)

	go func() {
		for d := range amqpd {
			deliveries <- &Delivery{&d}
		}

		close(deliveries)
	}()

	return deliveries, nil
}

func (ch *Channel) ExchangeDeclare(name, kind string, opt amqputil.Option) error {
	var (
		durable, autoDelete, internal, noWait bool
		args                                  amqp.Table
	)

	if v, ok := opt["durable"]; ok {
		durable, ok = v.(bool)

		if !ok {
			return errors.New("durable option is of type bool")
		}
	}

	if v, ok := opt["autoDelete"]; ok {
		autoDelete, ok = v.(bool)

		if !ok {
			return errors.New("autoDelete option is of type bool")
		}
	}

	if v, ok := opt["internal"]; ok {
		internal, ok = v.(bool)

		if !ok {
			return errors.New("internal option is of type bool")
		}
	}

	if v, ok := opt["noWait"]; ok {
		noWait, ok = v.(bool)

		if !ok {
			return errors.New("noWait option is of type bool")
		}
	}

	if v, ok := opt["args"]; ok {
		args, ok = v.(amqp.Table)

		if !ok {
			return errors.New("args is of type amqp.Table")
		}
	}

	return ch.Channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, args)
}

func (ch *Channel) QueueBind(name, key, exchange string, opt amqputil.Option) error {
	var (
		noWait bool
		args   amqp.Table
	)

	if v, ok := opt["noWait"]; ok {
		noWait, ok = v.(bool)

		if !ok {
			return errors.New("noWait option is of type bool")
		}
	}

	if v, ok := opt["args"]; ok {
		args, ok = v.(amqp.Table)

		if !ok {
			return errors.New("args is of type amqp.Table")
		}
	}

	return ch.Channel.QueueBind(name, key, exchange, noWait, args)
}

func (ch *Channel) QueueDeclare(name string, opt amqputil.Option) (amqputil.Queue, error) {
	var (
		durable, autoDelete, exclusive, noWait bool
		args                                   amqp.Table
	)

	if v, ok := opt["durable"]; ok {
		durable, ok = v.(bool)

		if !ok {
			return nil, errors.New("durable option is of type bool")
		}
	}

	if v, ok := opt["autoDelete"]; ok {
		autoDelete, ok = v.(bool)

		if !ok {
			return nil, errors.New("autoDelete option is of type bool")
		}
	}

	if v, ok := opt["exclusive"]; ok {
		exclusive, ok = v.(bool)

		if !ok {
			return nil, errors.New("Exclusive option is of type bool")
		}
	}

	if v, ok := opt["noWait"]; ok {
		noWait, ok = v.(bool)

		if !ok {
			return nil, errors.New("noWait option is of type bool")
		}
	}

	if v, ok := opt["args"]; ok {
		args, ok = v.(amqp.Table)

		if !ok {
			return nil, errors.New("args is of type amqp.Table")
		}
	}

	q, err := ch.Channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, args)

	if err != nil {
		return nil, err
	}

	return &Queue{&q}, nil
}
