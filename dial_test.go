package amqputil

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/fsouza/go-dockerclient"
)

var (
	dockerClient *docker.Client
)

func TestMain(m *testing.M) {
	var err error

	// Start the amqp backend on test startup
	endpoint := "unix:///var/run/docker.sock"
	dockerClient, err = docker.NewClient(endpoint)

	if err != nil {
		fmt.Printf("Error connecting to docker daemon...")
		os.Exit(1)
	}

	// remove any previously executed rabbitmq container with name
	// rabbitmq-tests
	rmRabbit()

	// Execute the tests
	status := m.Run()

	// remove the backend container
	rmRabbit()

	os.Exit(status)
}
func rmRabbit() {
	rmCtnOpt := docker.RemoveContainerOptions{
		ID:            "rabbitmq-tests",
		RemoveVolumes: true,
		Force:         true,
	}

	_ = dockerClient.RemoveContainer(rmCtnOpt)
}

func pullRabbit() error {
	pullOpts := docker.PullImageOptions{
		Repository:   "rabbitmq",
		Tag:          "latest",
		OutputStream: os.Stdout,
	}

	err := dockerClient.PullImage(pullOpts, docker.AuthConfiguration{})

	if err != nil {
		return err
	}

	return nil
}

// runRabbit execute a container with rabbitmq daemon
// If docker fail to create the container, then try to pull the
// last version of rabbitmq image from registry. If fail again,
// then returns error.
// When rabbitmq is successfully started, block until it finish the
// setup and is ready to accept connections.
func runRabbit() (*docker.Container, error) {
	var rabbitCtn *docker.Container

	createContainer := func() (*docker.Container, error) {
		opts := docker.CreateContainerOptions{
			Name: "rabbitmq-tests",
			Config: &docker.Config{
				Image: "rabbitmq",
				ExposedPorts: map[docker.Port]struct{}{
					"5672/tcp": {},
				},
			},
			HostConfig: &docker.HostConfig{
				PortBindings: map[docker.Port][]docker.PortBinding{
					"5672/tcp": []docker.PortBinding{docker.PortBinding{HostPort: "5672"}}},
				PublishAllPorts: true,
				Privileged:      false,
			},
		}

		rabbitCtn, err := dockerClient.CreateContainer(opts)

		return rabbitCtn, err
	}

	rabbitCtn, err := createContainer()

	if err != nil {
		err = pullRabbit()

		if err != nil {
			return nil, err
		}

		rabbitCtn, err = createContainer()

		if err != nil {
			return nil, err
		}
	}

	err = dockerClient.StartContainer(rabbitCtn.ID, nil)

	if err != nil {
		return nil, err
	}

	// block until rabbitmq can accept connections on port 5672
	waitRabbit()

	return rabbitCtn, nil
}

// waitRabbit blocks until rabbitmq can accept connections on
// localhost:5672
func waitRabbit() {
dial:
	conn, err := net.Dial("tcp", "localhost:5672")
	if err != nil {
		time.Sleep(500 * time.Millisecond)
		goto dial
	}

	fmt.Fprintf(conn, "AMQP%00091")
	_, err = bufio.NewReader(conn).ReadString('\n')

	if err != nil && err.Error() != "EOF" {
		conn.Close()
		time.Sleep(500 * time.Millisecond)
		goto dial
	}
}

// TestDial test a simple connection to rabbitmq.
// If the rabbitmq disconnects will not be tested here!
func TestDial(t *testing.T) {
	// Should fail
	conn := New()
	err := conn.Dial("amqp://guest:guest@localhost:5672/%2f")

	if err == nil {
		t.Error("No backend started... Should fail")
		return
	}

	rabbitCtn, err := runRabbit()
	if err != nil {
		t.Error(err)
	}

	err = conn.Dial("amqp://guest:guest@localhost:5672/%2f")

	if err != nil {
		t.Error(err)
		return
	}

	var success = false
	redialErrors := make(chan error)
	conn.AutoRedial(redialErrors, func () {
		// this function should be executed is successfully reconnected
		// here we can recover the application state (if needed)
		success = true
	})

	// required goroutine to consume connection error messages
	go func() {
		for {
			// discards the connection errors
			<-redialErrors
		}
	}()

	dockerClient.StopContainer(rabbitCtn.ID, 3)

	// concurrently starts the rabbitmq after 1 second
	go func() {
		time.Sleep(1 * time.Second)

		fmt.Printf("Starting rabbit\n")
		err := dockerClient.StartContainer(rabbitCtn.ID, nil)

		if err != nil {
			t.Error(err)
			return
		}
	}()

	// Wait 3 seconds to reconnect
	for i := 0; i < 3; i++ {
		runtime.Gosched()
		if conn.IsConnected {
			break
		}

		time.Sleep(time.Second)
	}

	if !conn.IsConnected || success == false {
		t.Errorf("Client doesn't reconnect in 3 seconds: %s", conn.IsConnected)
		return
	}

	conn.Conn.Close()
}
