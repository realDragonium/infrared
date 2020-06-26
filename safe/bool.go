package safe

import (
	"sync"
)

type Bool struct {
	sync.RWMutex
	Value bool
}

func (safe *Bool) Read() bool {
	safe.RLock()
	defer safe.RUnlock()
	return safe.Value
}

func (safe *Bool) Write(value bool) {
	safe.Lock()
	defer safe.Unlock()
	safe.Value = value
}
