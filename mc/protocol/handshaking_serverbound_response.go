package protocol

import pk "github.com/haveachin/infrared/mc/packet"

const HandshakingServerBoundResponsePacketID byte = 0x00

type HandshakingServerBoundResponse struct {
	JSONResponse pk.String
}

func (response HandshakingServerBoundResponse) Marshal() pk.Packet {
	return pk.Marshal(
		HandshakingServerBoundResponsePacketID,
		response.JSONResponse,
	)
}
