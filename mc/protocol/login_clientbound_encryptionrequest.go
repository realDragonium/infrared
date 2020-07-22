package protocol

import pk "github.com/haveachin/infrared/mc/packet"

const LoginClientBoundEncryptionRequestPacketID byte = 0x01

type LoginClientBoundEncryptionRequest struct {
	ServerID    pk.String
	PublicKey   pk.ByteArray
	VerifyToken pk.ByteArray
}

func (packet LoginClientBoundEncryptionRequest) Marshal() pk.Packet {
	return pk.Marshal(
		LoginClientBoundEncryptionRequestPacketID,
		packet.ServerID,
		packet.PublicKey,
		packet.VerifyToken,
	)
}
