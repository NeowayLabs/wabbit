package amqputil

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/fsouza/go-dockerclient"
)

var (
	dockerClient *docker.Client
)

func TestMain(t *testing.T) {
	var err error

	endpoint := "unix:///var/run/docker.sock"
	dockerClient, err = docker.NewClient(endpoint)

	if err != nil {
		t.Error(err)
		return
	}
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

func runRabbitMQ() (*docker.Container, error) {
	var pulled bool

	rmCtnOpt := docker.RemoveContainerOptions{
		ID: "rabbitmq-tests",
		RemoveVolumes: true,
		Force: true,
	}

	_ = dockerClient.RemoveContainer(rmCtnOpt)

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
	rabbitCtn, err := runRabbitMQ()

	if err != nil {
		t.Error(err)
		fmt.Println("%+v\n", reflect.TypeOf(err))
	}

	fmt.Println(rabbitCtn)
}
