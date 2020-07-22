package zlib

import (
	"github.com/4kills/zlib"
	"sync"
)

var (
	decoder    *zlib.Reader
	decodeLock sync.Mutex

	encoder    *zlib.Writer
	encodeLock sync.Mutex
)

func init() {
	var err error
	decoder, err = zlib.NewReader(nil)
	if err != nil {
		panic(err)
	}

	encoder = zlib.NewWriter(nil)
}

func Decode(in []byte) ([]byte, error) {
	decodeLock.Lock()
	defer decodeLock.Unlock()
	_, bb, err := decoder.ReadBytes(in)
	return bb, err
}

func Encode(in []byte) ([]byte, error) {
	encodeLock.Lock()
	defer encodeLock.Unlock()
	return encoder.WriteBytes(in)
}
