package mock

import (
	"errors"
	"sync"
)

var (
	servers map[string]*AMQPServer
	mu      *sync.Mutex
)

func init() {
	servers = make(map[string]*AMQPServer)

	mu = &sync.Mutex{}
}

// AMQPServer is a fake AMQP server
type AMQPServer struct {
	mu      *sync.Mutex
	running bool
	amqpuri string
}

// NewServer returns a new fake amqp server
func newServer(amqpuri string) *AMQPServer {
	return &AMQPServer{
		mu:      &sync.Mutex{},
		amqpuri: amqpuri,
	}
}

// Start a new AMQP server fake-listening on host:port
func (s *AMQPServer) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = true
	return nil
}

// Stop the fake server
func (s *AMQPServer) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.running = false
	return nil
}

// NewServer starts a new fake server
func NewServer(amqpuri string) *AMQPServer {
	var amqpServer *AMQPServer

	mu.Lock()
	defer mu.Unlock()

	amqpServer = servers[amqpuri]
	if amqpServer == nil {
		amqpServer = newServer(amqpuri)
		servers[amqpuri] = amqpServer
	}

	return amqpServer
}

func connect(amqpuri string) (*AMQPServer, error) {
	mu.Lock()
	defer mu.Unlock()

	amqpServer := servers[amqpuri]

	if amqpServer == nil {
		return nil, errors.New("Network unreachable")
	}

	return amqpServer, nil
}
