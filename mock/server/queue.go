package server

import "github.com/tiago4orion/amqputil"

type Queue struct {
	name string
	data chan amqputil.Delivery
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
