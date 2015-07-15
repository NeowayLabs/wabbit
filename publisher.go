package amqputil

import "github.com/streadway/amqp"

type Publisher struct {
	conn    *Conn
	channel *amqp.Channel
}

func NewPublisher(conn *Conn) (*Publisher, error) {
	pb := Publisher{
		conn: conn,
	}

	ch, err := conn.Channel()

	if err != nil {
		return nil, err
	}

	pb.channel = ch
	return &pb, nil
}

func (pb *Publisher) Publish(exc string, route string, message string) error {
	err := pb.channel.Publish(
		exc,   // publish to an exchange
		route, // routing to 0 or more queues
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(message),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	)

	return err
}
