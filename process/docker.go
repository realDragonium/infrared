package process

import (
	"errors"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
	"time"
)

const contextTimeout = 5 * time.Second

type DockerConfig struct {
	Address       string
	DNSServer     string
	ContainerName string
	Hibernate     bool
}

func (cfg DockerConfig) Validate() error {
	if cfg.ContainerName == "" {
		return errors.New("config docker: 'ContainerName' not set")
	}

	return nil
}

type docker struct {
	client        *client.Client
	containerName string
}

// NewDocker create a new docker process that manages a container
func NewDocker(cfg DockerConfig) (Process, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	var cli *client.Client
	var err error

	if cfg.Address == "" {
		cli, err = client.NewClientWithOpts(client.FromEnv)
	} else {
		cli, err = client.NewClientWithOpts(
			client.WithHost(cfg.Address),
			client.WithAPIVersionNegotiation(),
		)
	}

	if err != nil {
		return nil, err
	}

	return docker{
		client:        cli,
		containerName: cfg.ContainerName,
	}, nil
}

func (dkr docker) Start() error {
	containerID, err := dkr.resolveContainerName()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	return dkr.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

func (dkr docker) Stop() error {
	containerID, err := dkr.resolveContainerName()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	return dkr.client.ContainerStop(ctx, containerID, nil)
}

func (dkr docker) IsRunning() (bool, error) {
	containerID, err := dkr.resolveContainerName()
	if err != nil {
		return false, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	info, err := dkr.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return false, err
	}

	return info.State.Running, nil
}

func (dkr docker) resolveContainerName() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), contextTimeout)
	defer cancel()

	containers, err := dkr.client.ContainerList(ctx, types.ContainerListOptions{All: true})
	if err != nil {
		return "", err
	}

	for _, container := range containers {
		for _, name := range container.Names {
			if name != dkr.containerName {
				continue
			}
			return container.ID, nil
		}
	}

	return "", fmt.Errorf("container with name %s not found", dkr.containerName)
}
