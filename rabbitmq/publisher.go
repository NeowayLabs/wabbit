package rabbitmq

import "github.com/tiago4orion/amqputil"

type Publisher struct {
	conn    *Conn
	channel amqputil.Publisher
}

func NewPublisher(conn *Conn) (*Publisher, error) {
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

func (pb *Publisher) Publish(exc string, route string, message string) error {
	err := pb.channel.Publish(
		exc,   // publish to an exchange
		route, // routing to 0 or more queues
		[]byte(message),
	)

	return err
}
