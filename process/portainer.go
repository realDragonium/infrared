package process

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"io/ioutil"
	"net/http"
)

const (
	contentType            = "application/json"
	authenticationEndpoint = "http://%s/api/auth"
	dockerEndpoint         = "tcp://%s/api/endpoints/%s/docker"
)

type portainer struct {
	docker   docker
	address  string
	username string
	password string
	header   map[string]string
}

// NewPortainer creates a new portainer process that manages a docker container
func NewPortainer(containerName, address, endpointID, username, password string) (Process, error) {
	baseURL := fmt.Sprintf(dockerEndpoint, address, endpointID)
	header := map[string]string{}
	cli, err := client.NewClientWithOpts(
		client.WithHost(baseURL),
		client.WithScheme("http"),
		client.WithAPIVersionNegotiation(),
		client.WithHTTPHeaders(header),
	)
	if err != nil {
		return nil, err
	}

	return portainer{
		docker: docker{
			client:        cli,
			containerName: "/" + containerName,
		},
		address:  address,
		username: username,
		password: password,
		header:   header,
	}, nil
}

func (port portainer) Start() error {
	err := port.docker.Start()
	if err == nil {
		return nil
	}

	if !isUnauthorized(err) {
		return err
	}

	if err := port.authenticate(); err != nil {
		return fmt.Errorf("could not authorize; %s", err)
	}

	return port.docker.Start()
}

func (port portainer) Stop() error {
	err := port.docker.Stop()
	if err == nil {
		return nil
	}

	if !isUnauthorized(err) {
		return err
	}

	if err := port.authenticate(); err != nil {
		return fmt.Errorf("could not authorize; %s", err)
	}

	return port.docker.Stop()
}

func (port portainer) IsRunning() (bool, error) {
	isRunning, err := port.docker.IsRunning()
	if err == nil {
		return isRunning, nil
	}

	if !isUnauthorized(err) {
		return false, err
	}

	if err := port.authenticate(); err != nil {
		return false, fmt.Errorf("could not authorize; %s", err)
	}

	return port.docker.IsRunning()
}

func isUnauthorized(err error) bool {
	return errdefs.GetHTTPErrorStatusCode(err) == http.StatusUnauthorized
}

func (port *portainer) authenticate() error {
	var credentials = struct {
		Username string `json:"Username"`
		Password string `json:"Password"`
	}{
		Username: port.username,
		Password: port.password,
	}

	bodyJSON, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	url := fmt.Sprintf(authenticationEndpoint, port.address)
	response, err := http.Post(url, contentType, bytes.NewBuffer(bodyJSON))
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return errors.New(http.StatusText(response.StatusCode))
	}

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var jwtResponse = struct {
		JWT string `json:"jwt"`
	}{}

	if err := json.Unmarshal(data, &jwtResponse); err != nil {
		return err
	}

	port.header["Authorization"] = fmt.Sprintf("Bearer %s", jwtResponse.JWT)
	return nil
}
