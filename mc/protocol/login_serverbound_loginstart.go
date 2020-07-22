package protocol

import pk "github.com/haveachin/infrared/mc/packet"

const LoginServerBoundLoginStartPacketID byte = 0x00

type LoginServerLoginStart struct {
	Name pk.String
}

func ParseLoginServerBoundLoginStart(packet pk.Packet) (LoginServerLoginStart, error) {
	var start LoginServerLoginStart

	if packet.ID != LoginServerBoundLoginStartPacketID {
		return start, ErrInvalidPacketID
	}

	if err := packet.Scan(&start.Name); err != nil {
		return start, err
	}

	return start, nil
}
