package rabbitmq

import "github.com/tiago4orion/amqputil"

type Publisher struct {
	conn    amqputil.Conn
	channel amqputil.Publisher
}

func NewPublisher(conn amqputil.Conn) (*Publisher, error) {
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

func (pb *Publisher) Publish(exc string, route string, message string, opt amqputil.Option) error {
	err := pb.channel.Publish(
		exc,   // publish to an exchange
		route, // routing to 0 or more queues
		[]byte(message),
		opt,
	)

	return err
}
