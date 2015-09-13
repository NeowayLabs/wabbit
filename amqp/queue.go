package amqp

import "github.com/streadway/amqp"

type Queue struct {
	*amqp.Queue
}

func (q *Queue) Messages() int {
	return q.Queue.Messages
}

func (q *Queue) Name() string {
	return q.Queue.Name
}

func (q *Queue) Consumers() int {
	return q.Queue.Consumers
}
