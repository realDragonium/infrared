package packet

import "io"

type PeekReader interface {
	Peek(n int) ([]byte, error)
	io.Reader
}

type bytePeeker struct {
	PeekReader
	cursor int
}

func (p *bytePeeker) Read(b []byte) (int, error) {
	buf, err := p.Peek(len(b) + p.cursor)
	if err != nil {
		return 0, err
	}

	for i := 0; i < len(b); i++ {
		b[i] = buf[i+p.cursor]
	}

	p.cursor += len(b)

	return len(buf), nil
}

func (p *bytePeeker) ReadByte() (byte, error) {
	buf, err := p.Peek(1 + p.cursor)
	if err != nil {
		return 0x00, err
	}

	b := buf[p.cursor]
	p.cursor++

	return b, nil
}
