package process

type Config struct {
	System struct {
		Directory    string
		StartCommand string
		StopCommand  string
	}
	Docker struct {
		Address       string
		DNSServer     string
		ContainerName string
		Hibernate     bool
		Portainer     struct {
			EndpointID string
			Username   string
			Password   string
		}
	}
}

func (cfg Config) hasSystemConfig() bool {
	if cfg.System.StartCommand == "" {
		return false
	}

	return true
}

func (cfg Config) hasDockerConfig() bool {
	if cfg.Docker.ContainerName == "" {
		return false
	}

	return true
}

func (cfg Config) hasPortainerConfig() bool {
	if cfg.Docker.Address == "" {
		return false
	}

	if cfg.Docker.Portainer.EndpointID == "" {
		return false
	}

	if cfg.Docker.Portainer.Username == "" {
		return false
	}

	if cfg.Docker.Portainer.Password == "" {
		return false
	}

	return true
}
