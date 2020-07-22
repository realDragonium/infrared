package protocol

import (
	pk "github.com/haveachin/infrared/mc/packet"
)

const PlayClientBoundChatMessagePacketID byte = 0x0E

type PlayClientBoundChatMessage struct {
	JSONData pk.Chat
	Position pk.Byte
	Sender   pk.UUID
}

func UnmarshalPlayClientChatMessage(packet pk.Packet) (PlayClientBoundChatMessage, error) {
	var chatMessage PlayClientBoundChatMessage

	if packet.ID != PlayClientBoundChatMessagePacketID {
		return PlayClientBoundChatMessage{}, ErrInvalidPacketID
	}

	if err := packet.Scan(
		&chatMessage.JSONData,
		&chatMessage.Position,
		&chatMessage.Sender,
	); err != nil {
		return PlayClientBoundChatMessage{}, err
	}

	return chatMessage, nil
}

func (chatMessage PlayClientBoundChatMessage) Marshal() pk.Packet {
	return pk.Marshal(
		PlayClientBoundChatMessagePacketID,
		chatMessage.JSONData,
		chatMessage.Position,
		chatMessage.Sender,
	)
}
