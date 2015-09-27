package amqp

import "github.com/tiago4orion/wabbit"

type Publisher struct {
	conn    wabbit.Conn
	channel wabbit.Publisher
}

func NewPublisher(conn wabbit.Conn) (*Publisher, error) {
	var err error

	pb := Publisher{
		conn: conn,
	}

	pb.channel, err = conn.Channel()

	if err != nil {
		return nil, err
	}

	return &pb, nil
}

func (pb *Publisher) Publish(exc string, route string, message []byte, opt wabbit.Option) error {
	err := pb.channel.Publish(
		exc,   // publish to an exchange
		route, // routing to 0 or more queues
		message,
		opt,
	)

	return err
}
