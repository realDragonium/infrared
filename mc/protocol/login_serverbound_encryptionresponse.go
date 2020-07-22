package protocol

import pk "github.com/haveachin/infrared/mc/packet"

const LoginServerBoundEncryptionResponsePacketID = 0x01

type LoginServerBoundEncryptionResponse struct {
	SharedSecret pk.ByteArray
	VerifyToken  pk.ByteArray
}

func ParseLoginServerBoundEncryptionResponse(packet pk.Packet) (LoginServerBoundEncryptionResponse, error) {
	var encryptionResponse LoginServerBoundEncryptionResponse

	if packet.ID != LoginServerBoundEncryptionResponsePacketID {
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
