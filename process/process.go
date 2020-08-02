package process

// Process is an arbitrary process that can be started or stopped
type Process interface {
	Start() error
	Stop() error
	IsRunning() (bool, error)
}

type Config struct {
	Docker    DockerConfig
	Portainer PortainerConfig
	System    SystemConfig
}

func New(cfg Config) (Process, error) {
	err := cfg.Docker.Validate()

	if err == nil {
		if err := cfg.Portainer.Validate(); err == nil {
			return NewPortainer(cfg.Docker, cfg.Portainer)
		}

		return NewDocker(cfg.Docker)
	}

	return nil, err
}
