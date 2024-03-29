package amqp

import amqp "github.com/rabbitmq/amqp091-go"

// Queue is a wrapper for "rabbitmq/amqp091-go".Queue but implementing
// the wabbit.Queue interface.
type Queue struct {
	*amqp.Queue
}

// Messages returns the count of messages not awaiting acknowledgment
func (q *Queue) Messages() int {
	return q.Queue.Messages
}

// Name of the queue
func (q *Queue) Name() string {
	return q.Queue.Name
}

// Consumers returns the amount of consumers of this queue
func (q *Queue) Consumers() int {
	return q.Queue.Consumers
}
