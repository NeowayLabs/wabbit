package server

import (
	"fmt"
	"sync"
)

type Exchange interface {
	route(route string, d *Delivery) error
	addBinding(route string, b *BindingsMap)
	delBinding(route string)
}

type BindingsMap struct {
	queue   *Queue
	headers map[string]string
}

type TopicExchange struct {
	name     string
	bindings map[string]*BindingsMap
	mu       *sync.RWMutex
}

func NewTopicExchange(name string) *TopicExchange {
	return &TopicExchange{
		name:     name,
		bindings: make(map[string]*BindingsMap),
		mu:       &sync.RWMutex{},
	}
}

func (t *TopicExchange) addBinding(route string, b *BindingsMap) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.bindings[route] = b
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
	bindings map[string]*BindingsMap
	mu       *sync.RWMutex
}

func NewDirectExchange(name string) *DirectExchange {
	return &DirectExchange{
		name:     name,
		bindings: make(map[string]*BindingsMap),
		mu:       &sync.RWMutex{},
	}
}

func (d *DirectExchange) addBinding(route string, b *BindingsMap) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.bindings == nil {
		d.bindings = make(map[string]*BindingsMap)
	}

	d.bindings[route] = b
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
	bindings map[string]*BindingsMap
	mu       *sync.RWMutex
}

func NewHeadersExchange(name string) *HeadersExchange {
	return &HeadersExchange{
		name:     name,
		bindings: make(map[string]*BindingsMap),
		mu:       &sync.RWMutex{},
	}
}

func (t *HeadersExchange) addBinding(route string, b *BindingsMap) {
	t.mu.Lock()
	defer t.mu.Unlock()
	bindingKey := route
	if bindingKey == "" {
		bindingKey = b.queue.name
	}

	t.bindings[bindingKey] = b
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
		if match, err := headersMatch(bindings, d); match {
			bindings.queue.data <- d
		} else if err != nil {
			return err
		}
	}

	// The headers doesnt match any attribute, then will be discarded
	return nil
}
