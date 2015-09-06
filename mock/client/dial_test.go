package client

import (
	"errors"
	"testing"
	"time"

	"github.com/tiago4orion/amqputil"
	"github.com/tiago4orion/amqputil/mock/server"
)

var rabbitmqPort = "35672"

// WaitOK blocks until rabbitmq can accept connections on
// <ctn ip address>:5672
func waitRabbitOK(amqpuri string) error {
	var err error
	var counter = 0
	conn := NewConn()
dial:
	if counter > 120 {
		panic("Impossible to connect to rabbitmq")
	}

	err = conn.Dial(amqpuri)
	if err != nil {
		time.Sleep(500 * time.Millisecond)
		counter++
		goto dial
	}

	return nil
}

// TestDial test a simple connection to rabbitmq.
// If the rabbitmq disconnects will not be tested here!
func TestDial(t *testing.T) {
	amqpuri := "amqp://guest:guest@localhost:35672/%2f"

	// Should fail
	conn := NewConn()
	err := conn.Dial(amqpuri)

	if err == nil {
		t.Error("No backend started... Should fail")
		return
	}

	server := server.NewServer(amqpuri)
	err = waitRabbitOK(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	err = conn.Dial(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	server.Stop()
}

func TestAutoRedial(t *testing.T) {
	var err error

	amqpuri := "amqp://guest:guest@localhost:35672/%2f"

	server := server.NewServer(amqpuri)

	defer server.Stop()

	err = waitRabbitOK(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	conn := NewConn()
	err = conn.Dial(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	defer conn.Close()
	redialErrors := make(chan error)

	done := make(chan bool)
	conn.AutoRedial(redialErrors, done)

	// required goroutine to consume connection error messages
	go func() {
		for {
			// discards the connection errors
			<-redialErrors
		}
	}()

	server.Stop()

	// concurrently starts the rabbitmq after 1 second
	go func() {
		time.Sleep(1 * time.Second)
		server.Start()
	}()

	select {
	case <-time.After(3 * time.Second):
		err = errors.New("Timeout exceeded. AMQP reconnect failed")
	case <-done:
		err = nil
	}

	if err != nil {
		t.Errorf("Client doesn't reconnect in 3 seconds: %s", err.Error())
		return
	}
}

func TestChannelMock(t *testing.T) {
	var channel amqputil.Channel

	// rabbitmq.Channel satisfies amqputil.Channel interface
	channel = new(Channel)

	if channel == nil {
		t.Error("Maybe amqputil.Channel interface does not mock amqp.Channel correctly")
	}
}

func TestConnMock(t *testing.T) {
	var conn amqputil.Conn

	// rabbitmq.Conn satisfies amqputil.Conn interface
	conn = NewConn()

	if conn == nil {
		t.Error("Maybe amqputil.Conn interface does not mock amqp.Conn correctly")
	}
}
