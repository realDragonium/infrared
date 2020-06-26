package safe

import (
	"github.com/haveachin/infrared/process"
	"sync"
)

type Process struct {
	sync.Mutex
	Value process.Process
}

func (safe *Process) Update(proc process.Process) {
	safe.Lock()
	defer safe.Unlock()
	safe.Value = proc
}

func (safe *Process) Start() error {
	safe.Lock()
	defer safe.Unlock()
	return safe.Value.Start()
}

func (safe *Process) Stop() error {
	safe.Lock()
	defer safe.Unlock()
	return safe.Value.Stop()
}

func (safe *Process) IsRunning() (bool, error) {
	safe.Lock()
	defer safe.Unlock()
	return safe.Value.IsRunning()
}
