package cfb8

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"testing"
)

func initEncrypter() (cipher.Stream, error) {
	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}
	return NewEncrypter(block, secret), nil
}

func initDecrypter() (cipher.Stream, error) {
	secret := make([]byte, 16)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return nil, err
	}
	return NewDecrypter(block, secret), nil
}

func BenchmarkEncrypt10000Bytes(b *testing.B) {
	c, err := initEncrypter()
	if err != nil {
		b.Error(err)
	}

	dst := make([]byte, 10000)
	src := make([]byte, 10000)
	if _, err := rand.Read(src); err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		c.XORKeyStream(dst, src)
	}
}

func BenchmarkDecrypt10000Bytes(b *testing.B) {
	c, err := initDecrypter()
	if err != nil {
		b.Error(err)
	}

	dst := make([]byte, 10000)
	src := make([]byte, 10000)
	if _, err := rand.Read(src); err != nil {
		b.Error(err)
	}

	for n := 0; n < b.N; n++ {
		c.XORKeyStream(dst, src)
	}
}
