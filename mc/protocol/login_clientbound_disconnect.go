package protocol

import pk "github.com/haveachin/infrared/mc/packet"

const LoginClientBoundDisconnectPacketID byte = 0x00

type LoginClientBoundDisconnect struct {
	Reason pk.Chat
}

func (disconnect LoginClientBoundDisconnect) Marshal() pk.Packet {
	return pk.Marshal(
		LoginClientBoundDisconnectPacketID,
		disconnect.Reason,
	)
}
