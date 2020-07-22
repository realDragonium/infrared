package protocol

import pk "github.com/haveachin/infrared/mc/packet"

const LoginClientBoundLoginSuccessPacketID byte = 0x02

type LoginClientBoundLoginSuccess struct {
	UUID     pk.UUID
	Username pk.String
}

func ParseServerLoginLoginSuccess(packet pk.Packet) (LoginClientBoundLoginSuccess, error) {
	var loginSuccess LoginClientBoundLoginSuccess

	if packet.ID != LoginClientBoundLoginSuccessPacketID {
		return loginSuccess, ErrInvalidPacketID
	}

	if err := packet.Scan(
		&loginSuccess.UUID,
		&loginSuccess.Username,
	); err != nil {
		return loginSuccess, err
	}

	return loginSuccess, nil
}

func (packet LoginClientBoundLoginSuccess) Marshal() pk.Packet {
	return pk.Marshal(
		LoginClientBoundLoginSuccessPacketID,
		packet.UUID,
		packet.Username,
	)
}
