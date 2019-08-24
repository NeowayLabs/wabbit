package server

import (
	"fmt"
	"sync"
)

type Exchange interface {
	route(route string, d *Delivery) error
	addBinding(route string, q *Queue)
	delBinding(route string)
}

type TopicExchange struct {
	name     string
	bindings map[string]*Queue
	mu       *sync.RWMutex
}

func NewTopicExchange(name string) *TopicExchange {
	return &TopicExchange{
		name:     name,
		bindings: make(map[string]*Queue),
		mu:       &sync.RWMutex{},
	}
}

func (t *TopicExchange) addBinding(route string, q *Queue) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bindings[route] = q
}

func (t *TopicExchange) delBinding(route string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.bindings, route)
}

func (t *TopicExchange) route(route string, d *Delivery) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for bname, q := range t.bindings {
		if topicMatch(bname, route) {
			q.data <- d
			return nil
		}
	}

	// The route doesnt match any binding, then will be discarded
	return nil
}

type DirectExchange struct {
	name     string
	bindings map[string]*Queue
	mu       *sync.RWMutex
}

func NewDirectExchange(name string) *DirectExchange {
	return &DirectExchange{
		name:     name,
		bindings: make(map[string]*Queue),
		mu:       &sync.RWMutex{},
	}
}

func (d *DirectExchange) addBinding(route string, q *Queue) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.bindings == nil {
		d.bindings = make(map[string]*Queue)
	}

	d.bindings[route] = q
}

func (d *DirectExchange) delBinding(route string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.bindings, route)
}

func (d *DirectExchange) route(route string, delivery *Delivery) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if q, ok := d.bindings[route]; ok {
		q.data <- delivery
		return nil
	}

	return fmt.Errorf("No bindings to route: %s", route)

}
