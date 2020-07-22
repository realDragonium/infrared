package protocol

import (
	pk "github.com/haveachin/infrared/mc/packet"
)

const PlayServerChatMessagePacketID = 0x03

type PlayServerChatMessage struct {
	Message pk.String
}

func UnmarshalPlayServerChatMessage(packet pk.Packet) (PlayServerChatMessage, error) {
	var chatMessage PlayServerChatMessage

	if packet.ID != PlayServerChatMessagePacketID {
		return PlayServerChatMessage{}, ErrInvalidPacketID
	}

	if err := packet.Scan(
		&chatMessage.Message,
	); err != nil {
		return PlayServerChatMessage{}, err
	}

	return chatMessage, nil
}

func (chatMessage PlayServerChatMessage) Marshal() pk.Packet {
	return pk.Marshal(
		PlayServerChatMessagePacketID,
		chatMessage.Message,
	)
}
