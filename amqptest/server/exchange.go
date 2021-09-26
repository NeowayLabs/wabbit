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

type BindingsMap struct {
	queue   *Queue
	headers map[string]string
}

type TopicExchange struct {
	name     string
	bindings map[string]BindingsMap
	mu       *sync.RWMutex
}

func NewTopicExchange(name string) *TopicExchange {
	return &TopicExchange{
		name:     name,
		bindings: make(map[string]BindingsMap),
		mu:       &sync.RWMutex{},
	}
}

func (t *TopicExchange) addBinding(route string, q *Queue) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bindings[route] = BindingsMap{q, nil}
}

func (t *TopicExchange) delBinding(route string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.bindings, route)
}

func (t *TopicExchange) route(route string, d *Delivery) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for bname, bindings := range t.bindings {
		if topicMatch(bname, route) {
			bindings.queue.data <- d
			return nil
		}
	}

	// The route doesnt match any binding, then will be discarded
	return nil
}

type DirectExchange struct {
	name     string
	bindings map[string]BindingsMap
	mu       *sync.RWMutex
}

func NewDirectExchange(name string) *DirectExchange {
	return &DirectExchange{
		name:     name,
		bindings: make(map[string]BindingsMap),
		mu:       &sync.RWMutex{},
	}
}

func (d *DirectExchange) addBinding(route string, q *Queue) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.bindings == nil {
		d.bindings = make(map[string]BindingsMap)
	}

	d.bindings[route] = BindingsMap{q, nil}
}

func (d *DirectExchange) delBinding(route string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.bindings, route)
}

func (d *DirectExchange) route(route string, delivery *Delivery) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if bindings, ok := d.bindings[route]; ok {
		bindings.queue.data <- delivery
		return nil
	}

	return fmt.Errorf("no bindings to route: %s", route)

}

type HeadersExchange struct {
	name     string
	bindings map[string]BindingsMap
	mu       *sync.RWMutex
}

func NewHeadersExchange(name string) *HeadersExchange {
	return &HeadersExchange{
		name:     name,
		bindings: make(map[string]BindingsMap),
		mu:       &sync.RWMutex{},
	}
}

func (t *HeadersExchange) addBinding(route string, q *Queue) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bindings[route] = BindingsMap{q, nil}
}

func (t *HeadersExchange) delBinding(route string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.bindings, route)
}

func (t *HeadersExchange) route(route string, d *Delivery) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	for _, bindings := range t.bindings {
		if headersMatch(bindings) {
			bindings.queue.data <- d
			return nil
		}
	}

	// The headers doesnt match any attribute, then will be discarded
	return nil
}
