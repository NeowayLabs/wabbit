package amqptest

import "github.com/tiago4orion/wabbit"

type Publisher struct {
	channel wabbit.Publisher
	conn    wabbit.Conn
}

func NewPublisher(conn wabbit.Conn) (*Publisher, error) {
	ch, err := conn.Channel()

	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

func (pb *Publisher) Publish(exc string, route string, message string, opt wabbit.Option) error {
	err := pb.channel.Publish(
		exc,   // publish to an exchange
		route, // routing to 0 or more queues
		[]byte(message),
		opt,
	)

	return err
}
