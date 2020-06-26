package safe

import "sync"

type String struct {
	sync.RWMutex
	Value string
}

func (safe *String) Read() string {
	safe.RLock()
	defer safe.RUnlock()
	return safe.Value
}

func (safe *String) Write(value string) {
	safe.Lock()
	defer safe.Unlock()
	safe.Value = value
}
