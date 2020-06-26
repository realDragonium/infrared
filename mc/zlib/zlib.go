package zlib

import (
	"encoding/json"
	"github.com/4kills/zlib"
	"io/ioutil"
	"sync"
)

var (
	decoder    *zlib.Reader
	decodeLock sync.Mutex

	encoder    *zlib.Writer
	encodeLock sync.Mutex
)

var enbytes [][]byte
var debytes [][]byte

func init() {
	var err error
	decoder, err = zlib.NewReader(nil)
	if err != nil {
		panic(err)
	}

	encoder, err = zlib.NewWriter(nil)
	if err != nil {
		panic(err)
	}
}

func Decode(in []byte) ([]byte, error) {
	decodeLock.Lock()
	defer decodeLock.Unlock()
	//enbytes = append(enbytes, in)
	decoder, err := zlib.NewReader(nil)
	if err != nil {
		panic(err)
	}
	defer decoder.Close()
	return decoder.ReadBytes(in)
}

func Encode(in []byte) ([]byte, error) {
	encodeLock.Lock()
	defer encodeLock.Unlock()
	//debytes = append(debytes, in)
	encoder, err := zlib.NewWriter(nil)
	if err != nil {
		panic(err)
	}
	defer encoder.Close()
	return encoder.WriteBytes(in)
}

func WriteJSON() {
	file, _ := json.MarshalIndent(enbytes, "", " ")
	_ = ioutil.WriteFile("encodedBytes.json", file, 0644)
	file, _ = json.MarshalIndent(debytes, "", " ")
	_ = ioutil.WriteFile("decodedBytes.json", file, 0644)
}
