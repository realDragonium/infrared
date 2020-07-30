package process

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

type SystemConfig struct {
	Directory    string
	StartCommand string
	StopCommand  string
}

func (cfg SystemConfig) Validate() error {
	if cfg.StartCommand == "" {
		return errors.New("config system: 'StartCommand' not set")
	}

	return nil
}

type system struct {
	startCmd *exec.Cmd
	stopCmd  *exec.Cmd
}

func NewSystem(cfg SystemConfig) (Process, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	startCmd := parseCommand(cfg.StartCommand)
	if startCmd == nil {
		return nil, errors.New("no startCommand script defined")
	}
	startCmd.Dir = cfg.Directory

	stopCmd := parseCommand(cfg.StopCommand)
	if stopCmd != nil {
		stopCmd.Dir = cfg.Directory
	}

	return &system{
		startCmd: startCmd,
		stopCmd:  stopCmd,
	}, nil
}

func parseCommand(cmd string) *exec.Cmd {
	if cmd == "" {
		return nil
	}

	cmdWithArgs := strings.Split(cmd, " ")

	if len(cmdWithArgs) > 1 {
		return exec.Command(cmdWithArgs[0], cmdWithArgs[1:]...)
	}

	return exec.Command(cmd)
}

func (sys *system) Start() error {
	return sys.startCmd.Start()
}

func (sys *system) Stop() error {
	if sys.startCmd.Process == nil {
		return nil
	}

	if sys.stopCmd == nil {
		if runtime.GOOS == "windows" {
			return exec.Command("taskkill", "/F", "/T", "/PID", fmt.Sprint(sys.startCmd.Process.Pid)).Run()
		}

		return sys.startCmd.Process.Signal(syscall.SIGQUIT)
	}

	return sys.stopCmd.Start()
}

func (sys system) IsRunning() (bool, error) {
	if sys.startCmd.Process == nil {
		return false, nil
	}

	return true, nil
}
