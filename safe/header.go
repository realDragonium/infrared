package safe

import (
	"sync"
)

type Header struct {
	sync.RWMutex
	Value map[string]string
}

func (safe *Header) Put(key, value string) {
	safe.Lock()
	defer safe.Unlock()
	safe.Value[key] = value
}
