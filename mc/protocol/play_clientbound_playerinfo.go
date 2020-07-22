package protocol

import (
	"bytes"
	pk "github.com/haveachin/infrared/mc/packet"
)

const PlayClientPlayerInfoPacketID = 0x33

type PlayClientPlayerInfo struct {
	Action          pk.VarInt
	NumberOfPlayers pk.VarInt
	Player          []PlayClientPlayerInfoPlayer
}

type PlayClientPlayerInfoPlayer struct {
	UUID               pk.UUID
	Name               pk.String
	NumberOfProperties pk.VarInt
	Property           []PlayClientPlayerInfoPlayerProperty
	Gamemode           pk.VarInt
	Ping               pk.VarInt
	HasDisplayName     pk.Boolean
	DisplayName        pk.Chat
}

type PlayClientPlayerInfoPlayerProperty struct {
	Name      pk.String
	Value     pk.String
	IsSigned  pk.Boolean
	Signature pk.String
}

func UnmarshalPlayClientPlayerInfo(packet pk.Packet) (PlayClientPlayerInfo, error) {
	var playerInfo PlayClientPlayerInfo

	if packet.ID != PlayClientPlayerInfoPacketID {
		return playerInfo, ErrInvalidPacketID
	}

	r := bytes.NewReader(packet.Data)

	if err := pk.Scan(r,
		&playerInfo.Action,
		&playerInfo.NumberOfPlayers,
	); err != nil {
		return playerInfo, err
	}

	playerInfo.Player = make([]PlayClientPlayerInfoPlayer, playerInfo.NumberOfPlayers)

	for i := 0; i < int(playerInfo.NumberOfPlayers); i++ {
		if err := pk.Scan(r,
			&playerInfo.Player[i].UUID,
		); err != nil {
			return playerInfo, err
		}

		switch playerInfo.Action {
		case 0:
			if err := pk.Scan(r,
				&playerInfo.Player[i].Name,
				&playerInfo.Player[i].NumberOfProperties,
			); err != nil {
				return playerInfo, err
			}

			playerInfo.Player[i].Property = make([]PlayClientPlayerInfoPlayerProperty, playerInfo.Player[i].NumberOfProperties)

			for j := 0; j < int(playerInfo.Player[i].NumberOfProperties); j++ {
				if err := pk.Scan(r,
					&playerInfo.Player[i].Property[j].Name,
					&playerInfo.Player[i].Property[j].Value,
					&playerInfo.Player[i].Property[j].IsSigned,
				); err != nil {
					return playerInfo, err
				}

				if playerInfo.Player[i].Property[j].IsSigned {
					if err := pk.Scan(r,
						&playerInfo.Player[i].Property[j].Signature,
					); err != nil {
						return playerInfo, err
					}
				}
			}

			if err := pk.Scan(r,
				&playerInfo.Player[i].Gamemode,
				&playerInfo.Player[i].Ping,
				&playerInfo.Player[i].HasDisplayName,
			); err != nil {
				return playerInfo, err
			}

			if playerInfo.Player[i].HasDisplayName {
				if err := pk.Scan(r,
					&playerInfo.Player[i].DisplayName,
				); err != nil {
					return playerInfo, err
				}
			}
		case 1:
			if err := pk.Scan(r,
				&playerInfo.Player[i].Gamemode,
			); err != nil {
				return playerInfo, err
			}
		case 2:
			if err := pk.Scan(r,
				&playerInfo.Player[i].Ping,
			); err != nil {
				return playerInfo, err
			}
		case 3:
			if err := pk.Scan(r,
				&playerInfo.Player[i].HasDisplayName,
			); err != nil {
				return playerInfo, err
			}

			if playerInfo.Player[i].HasDisplayName {
				if err := pk.Scan(r,
					&playerInfo.Player[i].DisplayName,
				); err != nil {
					return playerInfo, err
				}
			}
		case 4:
		}
	}

	return playerInfo, nil
}

func (playerInfo PlayClientPlayerInfo) Marshal() pk.Packet {
	fields := []pk.FieldEncoder{
		playerInfo.Action,
		playerInfo.NumberOfPlayers,
	}

	for _, player := range playerInfo.Player {
		fields = append(
			fields,
			player.UUID,
		)

		switch playerInfo.Action {
		case 0:
			fields = append(
				fields,
				player.Name,
				player.NumberOfProperties,
			)

			for _, property := range player.Property {
				fields = append(
					fields,
					property.Name,
					property.Value,
					property.IsSigned,
				)

				if property.IsSigned {
					fields = append(
						fields,
						property.Signature,
					)
				}
			}

			fields = append(
				fields,
				player.Gamemode,
				player.Ping,
				player.HasDisplayName,
			)

			if player.HasDisplayName {
				fields = append(
					fields,
					player.DisplayName,
				)
			}
		case 1:
			fields = append(
				fields,
				player.Gamemode,
			)
		case 2:
			fields = append(
				fields,
				player.Ping,
			)
		case 3:
			fields = append(
				fields,
				player.HasDisplayName,
			)

			if player.HasDisplayName {
				fields = append(
					fields,
					player.DisplayName,
				)
			}
		case 4:
		}
	}

	return pk.Marshal(
		PlayClientPlayerInfoPacketID,
		fields...,
	)
}
