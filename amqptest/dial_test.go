package amqptest

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/NeowayLabs/wabbit"
	"github.com/NeowayLabs/wabbit/amqptest/server"
	"github.com/pborman/uuid"
)

var rabbitmqPort = "35672"

// WaitOK blocks until rabbitmq can accept connections on
// <ctn ip address>:5672
func waitRabbitOK(amqpuri string) error {
	var err error
	var counter = 0
dial:
	if counter > 120 {
		panic("Impossible to connect to rabbitmq")
	}

	_, err = Dial(amqpuri)
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
	conn, err := Dial(amqpuri)

	if err == nil {
		t.Error("No backend started... Should fail")
		return
	}

	server := server.NewServer(amqpuri)
	server.Start()

	err = waitRabbitOK(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	conn, err = Dial(amqpuri)

	if err != nil || conn == nil {
		t.Error(err)
		return
	}

	server.Stop()
}

func TestAutoRedial(t *testing.T) {
	var err error

	amqpuri := "amqp://guest:guest@localhost:35672/%2f"

	server := server.NewServer(amqpuri)
	server.Start()
	defer server.Stop()

	err = waitRabbitOK(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	conn, err := Dial(amqpuri)

	if err != nil {
		t.Error(err)
		return
	}

	defer conn.Close()
	redialErrors := make(chan wabbit.Error)

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
		time.Sleep(10 * time.Second)
		server.Start()
	}()

	select {
	case <-time.After(20 * time.Second):
		err = errors.New("Timeout exceeded. AMQP reconnect failed")
	case <-done:
		err = nil
	}

	if err != nil {
		t.Errorf("Client doesn't reconnect in 3 seconds: %s", err.Error())
		return
	}
}

func TestConnMock(t *testing.T) {
	var conn wabbit.Conn

	// rabbitmq.Conn satisfies wabbit.Conn interface
	conn = &Conn{}

	if conn == nil {
		t.Error("Maybe wabbit.Conn interface does not mock amqp.Conn correctly")
	}
}

func TestConnCloseBeforeServerStop(t *testing.T) {
	tests := []struct {
		name  string
		sleep time.Duration
	}{
		{
			name:  "10 ms sleep before server.Stop()",
			sleep: 10 * time.Millisecond,
		},
		{
			name:  "100 ms sleep before server.Stop()",
			sleep: 100 * time.Millisecond,
		},
		{
			name:  "1000 ms sleep before server.Stop()",
			sleep: time.Second,
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable.
		t.Run(tc.name, func(t *testing.T) {
			// Start a server.
			uri := fmt.Sprintf("amqp://guest:guest@localhost:35672/%s",
				uuid.New())
			server := server.NewServer(uri)
			server.Start()
			defer server.Stop()

			// Dial to the server.
			conn, err := Dial(uri)
			if err != nil {
				t.Fatalf("Dial(): %v", err)
			}
			conn.Close()

			// Sleep before stopping the server.
			time.Sleep(tc.sleep)
		})
	}
}
