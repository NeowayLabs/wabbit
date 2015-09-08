package server

import "github.com/tiago4orion/amqputil"

const (
	QueueMaxLen = 2 << 8
)

type Queue struct {
	name string
	data chan amqputil.Delivery
}

func NewQueue(name string) *Queue {
	return &Queue{
		name: name,
		data: make(chan amqputil.Delivery, QueueMaxLen),
	}
}

func (q *Queue) Consumers() int {
	return 0
}

func (q *Queue) Name() string {
	return q.name
}

func (q *Queue) Messages() int {
	return 0
}
