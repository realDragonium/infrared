package mots

import (
	"bytes"
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
	Src    PacketWriter
	Dst    PacketWriter
	Cancel bool
}

type InterceptFunc func(*Message)

func ChatMessageCommandMiddleware(msg *Message) {
	if msg.Author.IsNotClient() {
		return
	}

	chatMessage, err := protocol.UnmarshalPlayServerChatMessage(msg.Packet)
	if err != nil {
		return
	}

	if chatMessage.Message != "@infrared status" {
		return
	}

	msg.Cancel = true

	response := protocol.PlayClientBoundChatMessage{
		JSONData: pk.Chat(`{"text":"Infrared is up and running!","color":"dark_green"}`),
		Position: pk.Byte(2),
	}

	_ = msg.Src.WritePacket(response.Marshal())
}

func SpawnPlayerSkinMiddleware(players *safe.Players) func(*Message) {
	return func(msg *Message) {
		if msg.Author != AuthorServer {
			return
		}

		spawnPlayer, err := protocol.UnmarshalPlayClientSpawnPlayer(msg.Packet)
		if err != nil {
			return
		}

		for _, proxyPlayer := range players.Values() {
			if bytes.Compare(spawnPlayer.PlayerUUID[:], proxyPlayer.OfflineUUID[:]) != 0 {
				continue
			}

			spawnPlayer.PlayerUUID = proxyPlayer.UUID
			log.Printf("Spawning %s", proxyPlayer.Username)
			break
		}

		msg.Packet = spawnPlayer.Marshal()
	}
}

func PlayerInfoSkinMiddleware(players *safe.Players) func(*Message) {
	return func(msg *Message) {
		if msg.Author != AuthorServer {
			return
		}

		playerInfo, err := protocol.UnmarshalPlayClientPlayerInfo(msg.Packet)
		if err != nil {
			return
		}

		for n, player := range playerInfo.Player {
			for _, proxyPlayer := range players.Values() {
				if bytes.Compare(player.UUID[:], proxyPlayer.OfflineUUID[:]) != 0 {
					continue
				}

				player.UUID = proxyPlayer.UUID

				if playerInfo.Action != 0 {
					break
				}

				isSigned := proxyPlayer.SkinSignature != ""

				for m, property := range player.Property {
					if property.Name != "textures" {
						continue
					}

					property.Value = pk.String(proxyPlayer.Skin)
					property.IsSigned = pk.Boolean(isSigned)
					property.Signature = pk.String(proxyPlayer.SkinSignature)
					player.Property[m] = property
					log.Printf("Skin %s", proxyPlayer.Username)
					break
				}

				player.Property = append(
					player.Property,
					protocol.PlayClientPlayerInfoPlayerProperty{
						Name:      "textures",
						Value:     pk.String(proxyPlayer.Skin),
						IsSigned:  pk.Boolean(isSigned),
						Signature: pk.String(proxyPlayer.SkinSignature),
					},
				)
				player.NumberOfProperties++
				log.Printf("Skin %s", proxyPlayer.Username)
				break
			}
			playerInfo.Player[n] = player
		}

		msg.Packet = playerInfo.Marshal()
	}
}
