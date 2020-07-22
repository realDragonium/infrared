package protocol

import (
	pk "github.com/haveachin/infrared/mc/packet"
)

const PlayClientSpawnPlayerPacketID = 0x04

type PlayServerSpawnPlayer struct {
	EntityID   pk.VarInt
	PlayerUUID pk.UUID
	X          pk.Double
	Y          pk.Double
	Z          pk.Double
	Yaw        pk.Angle
	Pitch      pk.Angle
}

func UnmarshalPlayClientSpawnPlayer(packet pk.Packet) (PlayServerSpawnPlayer, error) {
	var spawnPlayer PlayServerSpawnPlayer

	if packet.ID != PlayClientSpawnPlayerPacketID {
		return spawnPlayer, ErrInvalidPacketID
	}

	if err := packet.Scan(
		&spawnPlayer.EntityID,
		&spawnPlayer.PlayerUUID,
		&spawnPlayer.X,
		&spawnPlayer.Y,
		&spawnPlayer.Z,
		&spawnPlayer.Yaw,
		&spawnPlayer.Pitch,
	); err != nil {
		return spawnPlayer, err
	}

	return spawnPlayer, nil
}

func (spawnPlayer PlayServerSpawnPlayer) Marshal() pk.Packet {
	return pk.Marshal(
		PlayClientSpawnPlayerPacketID,
		spawnPlayer.EntityID,
		spawnPlayer.PlayerUUID,
		spawnPlayer.X,
		spawnPlayer.Y,
		spawnPlayer.Z,
		spawnPlayer.Yaw,
		spawnPlayer.Pitch,
	)
}
