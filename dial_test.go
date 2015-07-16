package amqputil

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/tiago4orion/conjure"
)

var (
	rabbitctl *conjure.RabbitMQ
)

func TestMain(m *testing.M) {
	var err error

	ctnClient, err := conjure.NewClient()

	if err != nil || ctnClient == nil {
		panic(err.Error())
	}

	rabbitctl = conjure.NewRabbitMQ(ctnClient)

	// Execute the tests
	status := m.Run()

	// remove the backend container
	rabbitctl.Remove()
	os.Exit(status)
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

	err = rabbitctl.Run()
	if err != nil {
		t.Error(err)
	}

	err = rabbitctl.WaitOK("localhost")

	if err != nil {
		t.Error(err)
		return
	}

	err = conn.Dial("amqp://guest:guest@localhost:35672/%2f")

	if err != nil {
		t.Error(err)
		return
	}

	rabbitctl.Remove()
}

func TestAutoRedial(t *testing.T) {
	err :=rabbitctl.Run()

	if err != nil {
		t.Errorf("Failed to start rabbitmq: %s", err.Error())
		return
	}

	err = rabbitctl.WaitOK("localhost")

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

	rabbitctl.Stop()

	// concurrently starts the rabbitmq after 1 second
	go func() {
		time.Sleep(1 * time.Second)

		err := rabbitctl.Start()

		if err != nil {
			t.Error(err)
			return
		}
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
}
