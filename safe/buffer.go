package safe

import (
	"bytes"
	"sync"
)

type Buffer struct {
	sync.RWMutex
	Value *bytes.Buffer
}

func NewBuffer(buf []byte) *Buffer {
	return &Buffer{
		RWMutex: sync.RWMutex{},
		Value:   bytes.NewBuffer(buf),
	}
}

func (buffer *Buffer) Write(b []byte) (int, error) {
	buffer.Lock()
	defer buffer.Unlock()
	return buffer.Value.Write(b)
}

func (buffer *Buffer) Read(b []byte) (int, error) {
	buffer.Lock()
	defer buffer.Unlock()
	return buffer.Value.Read(b)
}

func (buffer *Buffer) ReadByte() (byte, error) {
	buffer.RLock()
	defer buffer.RUnlock()
	return buffer.Value.ReadByte()
}
