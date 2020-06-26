package safe

import (
	"sync"
	"time"
)

type Duration struct {
	sync.RWMutex
	Value time.Duration
}

func (safe *Duration) Read() time.Duration {
	safe.RLock()
	defer safe.RUnlock()
	return safe.Value
}

func (safe *Duration) Write(value time.Duration) {
	safe.Lock()
	defer safe.Unlock()
	safe.Value = value
}
