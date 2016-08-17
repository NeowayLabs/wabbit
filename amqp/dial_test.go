// +build integration

package amqp

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/NeowayLabs/wabbit"
	"github.com/fsouza/go-dockerclient"
	"github.com/tiago4orion/conjure"
)

var (
	dockerClient     *conjure.Client
	rabbitmqCtn      *docker.Container
	rabbitmqCtnName1 = "test-rabbitmq1"
	rabbitmqPort1    = "35672"
	rabbitmqSpec1    = `{
		"Name": "` + rabbitmqCtnName1 + `",
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
						"HostPort": "` + rabbitmqPort1 + `"
					}
				]
			},
			"PublishAllPorts": true,
			"Privileged": false
		}
	}`
	rabbitmqCtnName2 = "test-rabbitmq1"
	rabbitmqPort2    = "35673"
	rabbitmqSpec2    = `{
		"Name": "` + rabbitmqCtnName2 + `",
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
						"HostPort": "` + rabbitmqPort2 + `"
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
		os.Exit(1)
	}

	dockerClient.Remove(rabbitmqCtnName1)
	dockerClient.Remove(rabbitmqCtnName2)

	// Execute the tests
	status := m.Run()

	// remove the backend container
	dockerClient.Remove(rabbitmqCtnName1)
	dockerClient.Remove(rabbitmqCtnName2)
	os.Exit(status)
}

// WaitOK blocks until rabbitmq can accept connections on
// <ctn ip address>:5672
func waitRabbitOK(host string, port string) error {
	var err error
	var counter uint
dial:
	_, err = Dial("amqp://guest:guest@" + host + ":" + port + "/%2f")
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

	rabbitmqCtn, err = dockerClient.Run(rabbitmqSpec1)
	if err != nil {
		t.Error(err)
	}

	err = waitRabbitOK("localhost", rabbitmqPort1)

	if err != nil {
		t.Error(err)
		return
	}

	_, err = Dial("amqp://guest:guest@localhost:35672/%2f")

	if err != nil {
		t.Error(err)
		return
	}

	dockerClient.Remove(rabbitmqCtnName1)
}

func TestAutoRedial(t *testing.T) {
	var err error

	dockerClient.Remove(rabbitmqCtnName2)
	rabbitmqCtn, err = dockerClient.Run(rabbitmqSpec2)

	if err != nil {
		t.Errorf("Failed to start rabbitmq: %s", err.Error())
		return
	}

	err = waitRabbitOK("localhost", rabbitmqPort2)

	if err != nil {
		t.Error(err)
		return
	}

	conn, err := Dial("amqp://guest:guest@localhost:35673/%2f")

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

	dockerClient.StopContainer(rabbitmqCtnName2, 3)

	// concurrently starts the rabbitmq after 1 second
	go func() {
		time.Sleep(1 * time.Second)

		err := dockerClient.StartContainer(rabbitmqCtnName2, nil)

		if err != nil {
			t.Error(err)
			return
		}
	}()

	select {
	case <-time.After(10 * time.Second):
		err = errors.New("Failed to reconnect. Timeout exceeded")
	case <-done:
		err = nil
	}

	if err != nil {
		t.Errorf("Client doesn't reconnect in 10 seconds: %s", err.Error())
		return
	}

	conn.Close()
	dockerClient.Remove(rabbitmqCtnName2)
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
