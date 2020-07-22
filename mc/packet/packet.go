package packet

import (
	"bytes"
	"fmt"
	"github.com/haveachin/infrared/mc/zlib"
	"io"
)

// Packet define a net data package
type Packet struct {
	ID   byte
	Data []byte
}

func Parse(b []byte) Packet {
	return Packet{
		ID:   b[0],
		Data: b[1:],
	}
}

//Marshal generate Packet with the ID and Fields
func Marshal(ID byte, fields ...FieldEncoder) (pk Packet) {
	pk.ID = ID

	for _, v := range fields {
		pk.Data = append(pk.Data, v.Encode()...)
	}

	return
}

func Scan(r DecodeReader, fields ...FieldDecoder) error {
	for _, v := range fields {
		err := v.Decode(r)
		if err != nil {
			return err
		}
	}
	return nil
}

//Scan decode the packet and fill data into fields
func (p Packet) Scan(fields ...FieldDecoder) error {
	return Scan(bytes.NewReader(p.Data), fields...)
}

// Pack packs the packet and compresses it if it is larger then the given threshold
func (p *Packet) Pack(threshold int) ([]byte, error) {
	var packedData []byte
	data := []byte{p.ID}
	data = append(data, p.Data...)

	if threshold > 0 {
		if len(data) > threshold {
			length := VarInt(len(data)).Encode()
			var err error
			data, err = Compress(data)
			if err != nil {
				return nil, err
			}

			packedData = append(packedData, VarInt(len(length)+len(data)).Encode()...)
			packedData = append(packedData, length...)
		} else {
			packedData = append(packedData, VarInt(int32(len(data)+1)).Encode()...)
			packedData = append(packedData, 0x00)
		}
	} else {
		packedData = append(packedData, VarInt(int32(len(data))).Encode()...)
	}

	return append(packedData, data...), nil
}

func ReadRaw(r DecodeReader) ([]byte, error) {
	var packetLength VarInt
	if err := packetLength.Decode(r); err != nil {
		return nil, err
	}

	if packetLength < 1 {
		return nil, fmt.Errorf("packet length too short")
	}

	data := make([]byte, packetLength)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, fmt.Errorf("read content of packet fail: %v", err)
	}

	return data, nil
}

// RecvPacket receive a packet from server
func Read(r DecodeReader, isZlib bool) (Packet, error) {
	data, err := ReadRaw(r)
	if err != nil {
		return Packet{}, err
	}

	if isZlib {
		return Decompress(data)
	}

	return Parse(data), nil
}

func Peek(p PeekReader, zlib bool) (Packet, error) {
	r := bytePeeker{
		PeekReader: p,
		cursor:     0,
	}

	return Read(&r, zlib)
}

// Decompress
func Decompress(data []byte) (Packet, error) {
	reader := bytes.NewBuffer(data)

	var dataLength VarInt
	if err := dataLength.Decode(reader); err != nil {
		return Packet{}, err
	}

	decompressedData := make([]byte, dataLength)
	if dataLength != 0 { // != 0 means compressed, let's decompress
		var err error
		decompressedData, err = zlib.Decode(reader.Bytes())
		if err != nil {
			return Packet{}, err
		}
	} else {
		decompressedData = data[1:]
	}

	return Parse(decompressedData), nil
}

// Compress
func Compress(data []byte) ([]byte, error) {
	return zlib.Encode(data)
}
