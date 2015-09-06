package server

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

// AMQPServer is a fake AMQP server. It handle the fake TCP connection
type AMQPServer struct {
	mu      *sync.Mutex
	running bool
	amqpuri string

	vhost       VHost
	notifyChans map[string]chan error
}

// NewServer returns a new fake amqp server
func newServer(amqpuri string) *AMQPServer {
	return &AMQPServer{
		mu:          &sync.Mutex{},
		amqpuri:     amqpuri,
		notifyChans: make(map[string]chan error),
		vhost:       NewVHost("/"),
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

	for _, c := range s.notifyChans {
		c <- errors.New("Connection lost")
	}

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

func (s *AMQPServer) addNotify(connID string, nchan chan error) {
	s.notifyChans[connID] = nchan
}

func (s *AMQPServer) delNotify(connID string) {
	delete(s.notifyChans, connID)
}

func getServer(amqpuri string) (*AMQPServer, error) {
	mu.Lock()
	defer mu.Unlock()

	amqpServer := servers[amqpuri]

	if amqpServer == nil {
		return nil, errors.New("Network unreachable")
	}

	return amqpServer, nil
}

func Connect(amqpuri string, connID string, nchan chan error) (*AMQPServer, error) {
	amqpServer, err := getServer(amqpuri)

	if err != nil {
		return nil, err
	}

	amqpServer.addNotify(connID, nchan)
	return amqpServer, nil
}

func Close(amqpuri string, connID string) error {
	amqpServer, err := getServer(amqpuri)

	if err != nil {
		return errors.New("Failed to close connection")
	}

	amqpServer.delNotify(connID)
	return nil
}
