package mock

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/tiago4orion/amqputil"
)

var rabbitmqPort = "35672"

// WaitOK blocks until rabbitmq can accept connections on
// <ctn ip address>:5672
func waitRabbitOK(host string) error {
	var err error
	conn := New()
dial:
	err = conn.Dial("amqp://guest:guest@" + host + ":" + rabbitmqPort + "/%2f")
	if err != nil {
		fmt.Printf("Failed to connect to rabbitmq: %s\n", err.Error())
		time.Sleep(500 * time.Millisecond)
		goto dial
	}

	return nil
}

// TestDial test a simple connection to rabbitmq.
// If the rabbitmq disconnects will not be tested here!
func TestDial(t *testing.T) {
	// Should fail
	conn := New()
	err := conn.Dial("amqp://guest:guest@localhost:35672/%2f")

	if err == nil {
		t.Error("No backend started... Should fail")
		return
	}

	StartRabbitmq()

	err = waitRabbitOK("localhost")

	if err != nil {
		t.Error(err)
		return
	}

	err = conn.Dial("amqp://guest:guest@localhost:35672/%2f")

	if err != nil {
		t.Error(err)
		return
	}

	StopRabbitmq()
}

func TestAutoRedial(t *testing.T) {
	var err error

	StartRabbitmq()

	err = waitRabbitOK("localhost")

	if err != nil {
		t.Error(err)
		return
	}

	conn := New()
	err = conn.Dial("amqp://guest:guest@localhost:35672/%2f")

	if err != nil {
		t.Error(err)
		return
	}

	var success = false
	redialErrors := make(chan error)
	conn.AutoRedial(redialErrors, func() {
		// this function should be executed if successfully reconnected
		// here we can recover the application state after reconnect  (if needed)
		success = true
	})

	// required goroutine to consume connection error messages
	go func() {
		for {
			// discards the connection errors
			<-redialErrors
		}
	}()

	StopRabbitmq()

	// concurrently starts the rabbitmq after 1 second
	go func() {
		time.Sleep(1 * time.Second)

		StartRabbitmq()
	}()

	// Wait 3 seconds to reconnect
	for i := 0; i < 3; i++ {
		runtime.Gosched()
		if success {
			break
		}

		time.Sleep(time.Second)
	}

	if success == false {
		t.Errorf("Client doesn't reconnect in 3 seconds")
		return
	}

	conn.Close()

	StopRabbitmq()
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
	conn = New()

	if conn == nil {
		t.Error("Maybe amqputil.Conn interface does not mock amqp.Conn correctly")
	}
}
