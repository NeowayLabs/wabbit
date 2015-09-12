package server

import (
	"errors"
	"sync"

	"github.com/tiago4orion/wabbit"
	"github.com/tiago4orion/wabbit/amqptest"
	"github.com/tiago4orion/wabbit/utils"
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
	running bool
	amqpuri string

	vhost       *VHost
	notifyChans map[string]*utils.ErrBroadcast
}

// NewServer returns a new fake amqp server
func newServer(amqpuri string) *AMQPServer {
	return &AMQPServer{
		amqpuri:     amqpuri,
		notifyChans: make(map[string]*utils.ErrBroadcast),
		vhost:       NewVHost("/"),
	}
}

func (s *AMQPServer) CreateChannel() (wabbit.Channel, error) {
	return s.vhost, nil
}

// Start a new AMQP server fake-listening on host:port
func (s *AMQPServer) Start() error {
	mu.Lock()
	defer mu.Unlock()

	s.running = true
	return nil
}

// Stop the fake server
func (s *AMQPServer) Stop() error {
	mu.Lock()
	defer mu.Unlock()

	s.running = false

	for _, c := range s.notifyChans {
		c.Write(utils.NewError(
			amqptest.ChannelError,
			"channel/connection is not open",
			false,
			false,
		))
	}

	s.notifyChans = make(map[string]*utils.ErrBroadcast)
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

func (s *AMQPServer) addNotify(connID string, nchan *utils.ErrBroadcast) {
	mu.Lock()
	defer mu.Unlock()
	s.notifyChans[connID] = nchan
}

func (s *AMQPServer) delNotify(connID string) {
	mu.Lock()
	defer mu.Unlock()
	delete(s.notifyChans, connID)
}

func getServer(amqpuri string) (*AMQPServer, error) {
	mu.Lock()
	defer mu.Unlock()

	amqpServer := servers[amqpuri]

	if amqpServer == nil || amqpServer.running == false {
		return nil, errors.New("Network unreachable")
	}

	return amqpServer, nil
}

func Connect(amqpuri string, connID string, errBroadcast *utils.ErrBroadcast) (*AMQPServer, error) {
	amqpServer, err := getServer(amqpuri)

	if err != nil {
		return nil, err
	}

	amqpServer.addNotify(connID, errBroadcast)
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
