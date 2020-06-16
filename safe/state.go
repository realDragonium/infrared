package safe

import (
	"github.com/haveachin/infrared/mc/protocol"
	"sync"
)

type State struct {
	sync.RWMutex
	Value protocol.State
}

func (safe *State) Read() protocol.State {
	safe.RLock()
	defer safe.RUnlock()
	return safe.Value
}

func (safe *State) Write(state protocol.State) {
	safe.Lock()
	defer safe.Unlock()
	safe.Value = state
}
