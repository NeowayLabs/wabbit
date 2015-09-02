package mock

type Queue struct {
	name string
	consumers int
	messages int
}

func (q *Queue) Name() string {
	return q.name
}

func (q *Queue) Messages() int {
	return q.messages
}

func (q *Queue) Consumers() int {
	return q.consumers
}