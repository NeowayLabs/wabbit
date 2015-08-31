package rabbitmq

import (
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/fsouza/go-dockerclient"
	"github.com/tiago4orion/amqputil"
	"github.com/tiago4orion/conjure"
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
		panic(err.Error())
	}

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

	rabbitmqCtn, err = dockerClient.Run(rabbitmqSpec)
	if err != nil {
		t.Error(err)
	}

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

	dockerClient.Remove(rabbitmqCtnName)
}

func TestChannelMock(t *testing.T) {
	var channel amqputil.Channel

	channel = new(Channel)

	if channel == nil {
		t.Error("Maybe amqputil.Channel interface does not mock amqp.Channel correctly")
	}
}

func TestConnMock(t *testing.T) {
	var conn amqputil.Conn

	conn = New()

	if conn == nil {
		t.Error("Maybe amqputil.Conn interface does not mock amqp.Conn correctly")
	}
}
