package mock

import "github.com/tiago4orion/amqputil"

type Publisher struct {
	channel amqputil.Publisher
	conn    amqputil.Conn
}

func NewPublisher(conn amqputil.Conn) (*Publisher, error) {

	ch, err := conn.Channel()

	if err != nil {
		return nil, err
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

func (pb *Publisher) Publish(exc string, route string, message string) error {
	err := pb.channel.Publish(
		exc,   // publish to an exchange
		route, // routing to 0 or more queues
		[]byte(message),
	)

	return err
}
