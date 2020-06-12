package mots

import (
	pk "github.com/haveachin/infrared/mc/packet"
	"github.com/haveachin/infrared/mc/protocol"
	"github.com/haveachin/infrared/safe"
	"log"
)

type PacketWriter interface {
	WritePacket(pk.Packet) error
}

type Message struct {
	State  protocol.State
	Packet pk.Packet
	Author Author
	Dst    PacketWriter
}

type InterceptFunc func(*Message)

/*
ClientBound - Required by the client
ServerBound - Required by the server
Omnidirectional - Required by one, could also be both, idc
Bidirectional - Required by both
*/

func OmnidirectionalLogger(msg *Message) {
	/*if msg.State.IsPlay() {
		return
	}*/

	log.Println(msg.Author, msg.State, msg.Packet)
}

func BidirectionalStateUpdate(state *safe.State) InterceptFunc {
	return func(msg *Message) {
		switch state.Read() {
		case protocol.StateHandshaking:
			if !msg.Author.IsClient() {
				return
			}

			handshake, err := protocol.ParseSLPHandshake(msg.Packet)
			if err != nil {
				return
			}

			if handshake.IsStatusRequest() {
				state.Write(protocol.StateStatus)
			} else if handshake.IsLoginRequest() {
				state.Write(protocol.StateLogin)
			}
		case protocol.StateLogin:
			if !msg.Author.IsServer() {
				return
			}

			if msg.Packet.ID != protocol.ServerLoginLoginSuccessPacketID {
				return
			}

			state.Write(protocol.StatePlay)
		}
	}
}
