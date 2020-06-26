package process

import (
	"time"
)

const contextTimeout = 5 * time.Second

// Process is an arbitrary process that can be started or stopped
type Process interface {
	Start() error
	Stop() error
	IsRunning() (bool, error)
}

func New(cfg Config) (Process, error) {
	if cfg.hasSystemConfig() {
		return NewLocal(
			cfg.System.Directory,
			cfg.System.StartCommand,
			cfg.System.StopCommand,
		)
	}

	if cfg.hasPortainerConfig() {
		return NewPortainer(
			cfg.Docker.ContainerName,
			cfg.Docker.Address,
			cfg.Docker.Portainer.EndpointID,
			cfg.Docker.Portainer.Username,
			cfg.Docker.Portainer.Password,
		)
	}

	return NewDocker(
		cfg.Docker.Address,
		cfg.Docker.ContainerName,
	)
}
