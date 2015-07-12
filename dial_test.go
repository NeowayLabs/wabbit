package amqputil

import (
	"fmt"
	"os"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

var (
	dockerClient *docker.Client
)

func TestMain(m *testing.M) {
	var err error

	fmt.Printf("Starting backends")
	endpoint := "unix:///var/run/docker.sock"
	dockerClient, err = docker.NewClient(endpoint)

	if err != nil {
		fmt.Printf("Error connecting to docker daemon...")
		os.Exit(1)
	}

	rmRabbit()
	runRabbit()

	status := m.Run()

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

func runRabbit() (*docker.Container, error) {
	var pulled bool
try:
	opts := docker.CreateContainerOptions{
		Name: "rabbitmq-tests",
		Config: &docker.Config{
			Image: "rabbitmq",
			ExposedPorts: map[docker.Port]struct{}{
				"5674":  struct{}{},
				"15674": struct{}{},
			},
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
	if err != nil {
		t.Error(err)
	}

	fmt.Println(rabbitCtn)
}
