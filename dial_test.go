package amqputil

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/fsouza/go-dockerclient"
)

var (
	dockerClient *docker.Client
)

func TestMain(m *testing.M) {
	var err error

	fmt.Printf("Starting backends\n")
	endpoint := "unix:///var/run/docker.sock"
	dockerClient, err = docker.NewClient(endpoint)

	if err != nil {
		fmt.Printf("Error connecting to docker daemon...")
		os.Exit(1)
	}

	rmRabbit()

	status := m.Run()

	rmRabbit()

	os.Exit(status)
}
func rmRabbit() {
	fmt.Printf("Removing rabbitmq backend\n")
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

func runRabbit() (*docker.Container, error) {
	var pulled bool
try:
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

	if err != nil {
		if !pulled {
			err = pullRabbit()

			if err != nil {
				return nil, err
			}

			pulled = true
			goto try
		}

		return nil, err
	}

	err = dockerClient.StartContainer(rabbitCtn.ID, nil)

	if err != nil {
		return nil, err
	}

	return rabbitCtn, nil
}

func TestDial(t *testing.T) {
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

	// Wait a few seconds to rabbitmq finish the setup
	time.Sleep(3 * time.Second)

	err = conn.Dial("amqp://guest:guest@localhost:5672/%2f")

	if err != nil {
		t.Error(err)
		return
	}

	conn.AutoRedial()

	dockerClient.StopContainer(rabbitCtn.ID, 3)

	time.Sleep(10 * time.Second)

	err = dockerClient.StartContainer(rabbitCtn.ID, nil)

	if err != nil {
		t.Error(err)
		return
	}

	time.Sleep(3 * time.Second)

	

	fmt.Println(rabbitCtn)
}
