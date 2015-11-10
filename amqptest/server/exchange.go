package server

import "fmt"

type Exchange interface {
	route(route string, message []byte) (*Queue, error)
	addBinding(route string, q *Queue)
	delBinding(route string)
}

type TopicExchange struct {
	name     string
	bindings map[string]*Queue
}

func NewTopicExchange(name string) *TopicExchange {
	return &TopicExchange{
		name:     name,
		bindings: make(map[string]*Queue),
	}
}

func (t *TopicExchange) addBinding(route string, q *Queue) {
	t.bindings[route] = q
}

func (t *TopicExchange) delBinding(route string) {
	delete(t.bindings, route)
}

func (t *TopicExchange) route(route string, _ []byte) (*Queue, error) {
	for bname, q := range t.bindings {
		if topicMatch(bname, route) {
			return q, nil
		}
	}

	return nil, fmt.Errorf("Route '%s' doesn't match any routing-key", route)
}

type DirectExchange struct {
	name     string
	bindings map[string]*Queue
}

func NewDirectExchange(name string) *DirectExchange {
	return &DirectExchange{
		name:     name,
		bindings: make(map[string]*Queue),
	}
}

func (d *DirectExchange) addBinding(route string, q *Queue) {
	d.bindings[route] = q
}

func (d *DirectExchange) delBinding(route string) {
	delete(d.bindings, route)
}

func (d *DirectExchange) route(route string, _ []byte) (*Queue, error) {
	if q, ok := d.bindings[route]; ok {
		return q, nil
	}

	return nil, fmt.Errorf("No bindings to route: %s", route)

}
