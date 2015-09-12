package rabbitmq

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/tiago4orion/conjure"
	"github.com/tiago4orion/wabbit"
)

var (
	dockerClient    *conjure.Client
	rabbitmqCtn     *docker.Container
	rabbitmqCtnName = "test-rabbitmq"
	rabbitmqPort    = "35672"
	rabbitmqSpec    = `{
		"Name": "` + rabbitmqCtnName + `",
		"Config": {
			"Image": "rabbitmq",
			"ExposedPorts": {
				"5672/tcp": {}
			}
		},
		"HostConfig": {
			"PortBindings": {
				"5672/tcp": [
					{
						"HostPort": "` + rabbitmqPort + `"
					}
				]
			},
			"PublishAllPorts": true,
			"Privileged": false
		}
	}`
)

func TestMain(m *testing.M) {
	var err error

	dockerClient, err = conjure.NewClient()

	if err != nil || dockerClient == nil {
		fmt.Printf("You need docker >= 1.6 installed to enable testing rabbitmq backend\n")
		panic(err.Error())
	}

	dockerClient.Remove(rabbitmqCtnName)

	// Execute the tests
	status := m.Run()

	// remove the backend container
	dockerClient.Remove(rabbitmqCtnName)
	os.Exit(status)
}

// WaitOK blocks until rabbitmq can accept connections on
// <ctn ip address>:5672
func waitRabbitOK(host string) error {
	var err error
	var counter uint
dial:
	_, err = Dial("amqp://guest:guest@" + host + ":" + rabbitmqPort + "/%2f")
	if err != nil {
		if counter >= 120 {
			panic("isn't possible to connect on rabbitmq")
		}

		counter++
		time.Sleep(500 * time.Millisecond)
		goto dial
	}

	return nil
}

// TestDial test a simple connection to rabbitmq.
// If the rabbitmq disconnects will not be tested here!
func TestDial(t *testing.T) {
	// Should fail
	_, err := Dial("amqp://guest:guest@localhost:35672/%2f")

	if err == nil {
		t.Error("No backend started... Should fail")
		return
	}

	rabbitmqCtn, err = dockerClient.Run(rabbitmqSpec)
	if err != nil {
		t.Error(err)
	}

	err = waitRabbitOK("localhost")

	if err != nil {
		t.Error(err)
		return
	}

	_, err = Dial("amqp://guest:guest@localhost:35672/%2f")

	if err != nil {
		t.Error(err)
		return
	}

	dockerClient.Remove(rabbitmqCtnName)
}

func TestAutoRedial(t *testing.T) {
	var err error

	rabbitmqCtn, err = dockerClient.Run(rabbitmqSpec)

	if err != nil {
		t.Errorf("Failed to start rabbitmq: %s", err.Error())
		return
	}

	err = waitRabbitOK("localhost")

	if err != nil {
		t.Error(err)
		return
	}

	conn, err := Dial("amqp://guest:guest@localhost:35672/%2f")

	if err != nil {
		t.Error(err)
		return
	}

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

	dockerClient.StopContainer(rabbitmqCtnName, 3)

	// concurrently starts the rabbitmq after 1 second
	go func() {
		time.Sleep(1 * time.Second)

		err := dockerClient.StartContainer(rabbitmqCtnName, nil)

		if err != nil {
			t.Error(err)
			return
		}
	}()

	select {
	case <-time.After(5 * time.Second):
		err = errors.New("Failed to reconnect. Timeout exceeded")
	case <-done:
		err = nil
	}

	if err != nil {
		t.Errorf("Client doesn't reconnect in 3 seconds: %s", err.Error())
		return
	}

	conn.Close()
	dockerClient.Remove(rabbitmqCtnName)
}

func TestChannelMock(t *testing.T) {
	var channel wabbit.Channel

	// rabbitmq.Channel satisfies wabbit.Channel interface
	channel = new(Channel)

	if channel == nil {
		t.Error("Maybe wabbit.Channel interface does not mock amqp.Channel correctly")
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
