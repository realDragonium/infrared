package process

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
)

type system struct {
	startCmd *exec.Cmd
	stopCmd  *exec.Cmd
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

func NewLocal(directory string, startCommand string, stopCommand string) (Process, error) {
	startCmd := parseCommand(startCommand)
	if startCmd == nil {
		return nil, errors.New("no startCommand script defined")
	}
	startCmd.Dir = directory

	stopCmd := parseCommand(stopCommand)
	if stopCmd != nil {
		stopCmd.Dir = directory
	}

	return &system{
		startCmd: startCmd,
		stopCmd:  stopCmd,
	}, nil
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
