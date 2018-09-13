package server

import "github.com/PeriscopeData/wabbit"

const (
	QueueMaxLen = 2 << 8
)

type Queue struct {
	name string
	data chan wabbit.Delivery
}

func NewQueue(name string) *Queue {
	return &Queue{
		name: name,
		data: make(chan wabbit.Delivery, QueueMaxLen),
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
