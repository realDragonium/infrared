package protocol

import (
	pk "github.com/haveachin/infrared/mc/packet"
	"strings"
)

const (
	HandshakingServerBoundHandshakePacketID byte = 0x00

	HandshakingServerBoundHandshakeStatusState = pk.Byte(1)
	HandshakingServerBoundHandshakeLoginState  = pk.Byte(2)

	ForgeAddressSuffix  = "\x00FML\x00"
	Forge2AddressSuffix = "\x00FML2\x00"
)

type HandshakingServerBoundHandshake struct {
	ProtocolVersion pk.VarInt
	ServerAddress   pk.String
	ServerPort      pk.UnsignedShort
	NextState       pk.Byte
}

func ParseHandshakingServerBoundHandshake(packet pk.Packet) (HandshakingServerBoundHandshake, error) {
	var handshake HandshakingServerBoundHandshake

	if packet.ID != HandshakingServerBoundHandshakePacketID {
		return handshake, ErrInvalidPacketID
	}

	if err := packet.Scan(
		&handshake.ProtocolVersion,
		&handshake.ServerAddress,
		&handshake.ServerPort,
		&handshake.NextState); err != nil {
		return handshake, err
	}

	return handshake, nil
}

func (handshake HandshakingServerBoundHandshake) Marshal() pk.Packet {
	return pk.Marshal(
		HandshakingServerBoundHandshakePacketID,
		handshake.ProtocolVersion,
		handshake.ServerAddress,
		handshake.ServerPort,
		handshake.NextState)
}

func (handshake HandshakingServerBoundHandshake) IsStatusRequest() bool {
	return handshake.NextState == HandshakingServerBoundHandshakeStatusState
}

func (handshake HandshakingServerBoundHandshake) IsLoginRequest() bool {
	return handshake.NextState == HandshakingServerBoundHandshakeLoginState
}

func (handshake HandshakingServerBoundHandshake) IsForgeAddress() bool {
	addr := string(handshake.ServerAddress)

	if strings.HasSuffix(addr, ForgeAddressSuffix) {
		return true
	}

	if strings.HasSuffix(addr, Forge2AddressSuffix) {
		return true
	}

	return false
}

func (handshake HandshakingServerBoundHandshake) ParseServerAddress() string {
	addr := string(handshake.ServerAddress)
	addr = strings.TrimSuffix(addr, ForgeAddressSuffix)
	addr = strings.TrimSuffix(addr, Forge2AddressSuffix)
	addr = strings.Trim(addr, ".")
	return addr
}
