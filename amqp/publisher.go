package amqp

import (
	"github.com/NeowayLabs/wabbit"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn    wabbit.Conn
	channel wabbit.Publisher
}

func NewPublisher(conn wabbit.Conn, channel wabbit.Channel) (*Publisher, error) {
	var err error

	pb := Publisher{
		conn: conn,
	}

	if channel == nil {
		channel, err = conn.Channel()

		if err != nil {
			return nil, err
		}
	}

	pb.channel = channel

	return &pb, nil
}

func (pb *Publisher) Publish(
	exchange, key string,
	mandatory, immediate bool,
	msg amqp.Publishing) error {
	err := pb.channel.Publish(
		exchange, // publish to an exchange
		key,      // routing to 0 or more queues
		mandatory,
		immediate,
		msg,
	)

	return err
}
