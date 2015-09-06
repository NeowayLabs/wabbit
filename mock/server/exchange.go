package server

import "fmt"

type Exchange interface {
	route(route string, message []byte) (*Queue, error)
	addBinding(route string, q *Queue)
}

type TopicExchange struct {
	bindings map[string]*Queue
}

func (t *TopicExchange) addBinding(route string, q *Queue) {
	t.bindings[route] = q
}

func (t *TopicExchange) route(route string, message []byte) (*Queue, error) {
	for bname, q := range t.bindings {
		if topicMatch(bname, route) {
			return q, nil
		}
	}

	return nil, fmt.Errorf("Route '%s' doesn't match any routing-key", route)
}
