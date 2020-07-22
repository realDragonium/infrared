package protocol

import pk "github.com/haveachin/infrared/mc/packet"

const LoginClientBoundSetCompressionPacketID byte = 0x03

type LoginClientBoundSetCompression struct {
	Threshold pk.VarInt
}

func (packet LoginClientBoundSetCompression) Marshal() pk.Packet {
	return pk.Marshal(
		LoginClientBoundSetCompressionPacketID,
		packet.Threshold,
	)
}

func ParseLoginClientBoundSetCompression(packet pk.Packet) (LoginClientBoundSetCompression, error) {
	var setCompression LoginClientBoundSetCompression

	if packet.ID != LoginClientBoundSetCompressionPacketID {
		return setCompression, ErrInvalidPacketID
	}

	if err := packet.Scan(&setCompression.Threshold); err != nil {
		return setCompression, err
	}

	return setCompression, nil
}
