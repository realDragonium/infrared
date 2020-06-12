package protocol

import pk "github.com/haveachin/infrared/mc/packet"

const (
	ClientLoginStartPacketID              = 0x00
	ClientLoginEncryptionResponsePacketID = 0x01
)

const (
	ServerLoginDisconnectPacketID        = 0x00
	ServerLoginEncryptionRequestPacketID = 0x01
	ServerLoginLoginSuccessPacketID      = 0x02
	ServerLoginSetCompressionPacketID    = 0x03
)

type ClientLoginStart struct {
	Name pk.String
}

func ParseClientLoginStart(packet pk.Packet) (ClientLoginStart, error) {
	var start ClientLoginStart

	if packet.ID != ClientLoginStartPacketID {
		return start, ErrInvalidPacketID
	}

	if err := packet.Scan(&start.Name); err != nil {
		return start, err
	}

	return start, nil
}

type ClientLoginEncryptionResponse struct {
	SharedSecret pk.ByteArray
	VerifyToken  pk.ByteArray
}

func ParseClientLoginEncryptionResponse(packet pk.Packet) (ClientLoginEncryptionResponse, error) {
	var encryptionResponse ClientLoginEncryptionResponse

	if packet.ID != ClientLoginEncryptionResponsePacketID {
		return encryptionResponse, ErrInvalidPacketID
	}

	if err := packet.Scan(
		&encryptionResponse.SharedSecret,
		&encryptionResponse.VerifyToken,
	); err != nil {
		return encryptionResponse, err
	}

	return encryptionResponse, nil
}

type ServerLoginDisconnect struct {
	Reason pk.Chat
}

func (packet ServerLoginDisconnect) Marshal() pk.Packet {
	return pk.Marshal(
		ServerLoginDisconnectPacketID,
		packet.Reason,
	)
}

type ServerLoginEncryptionRequest struct {
	ServerID    pk.String
	PublicKey   pk.ByteArray
	VerifyToken pk.ByteArray
}

func (packet ServerLoginEncryptionRequest) ID() byte {
	return ServerLoginEncryptionRequestPacketID
}

func (packet ServerLoginEncryptionRequest) Marshal() pk.Packet {
	return pk.Marshal(
		ServerLoginEncryptionRequestPacketID,
		packet.ServerID,
		packet.PublicKey,
		packet.VerifyToken,
	)
}

type ServerLoginLoginSuccess struct {
	UUID     pk.String
	Username pk.String
}

func (packet ServerLoginLoginSuccess) Marshal() pk.Packet {
	return pk.Marshal(
		ServerLoginLoginSuccessPacketID,
		packet.UUID,
		packet.Username,
	)
}

type ServerLoginSetCompression struct {
	Threshold pk.VarInt
}

func (packet ServerLoginSetCompression) Marshal() pk.Packet {
	return pk.Marshal(
		ServerLoginSetCompressionPacketID,
		packet.Threshold,
	)
}

func ParseServerLoginSetCompression(packet pk.Packet) (ServerLoginSetCompression, error) {
	var setCompression ServerLoginSetCompression

	if packet.ID != ServerLoginSetCompressionPacketID {
		return setCompression, ErrInvalidPacketID
	}

	if err := packet.Scan(&setCompression.Threshold); err != nil {
		return setCompression, err
	}

	return setCompression, nil
}
