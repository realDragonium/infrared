package sim

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/haveachin/infrared/mc"
	"github.com/haveachin/infrared/mc/cfb8"
	pk "github.com/haveachin/infrared/mc/packet"
	"github.com/haveachin/infrared/mc/protocol"
	"github.com/haveachin/infrared/mc/sha1"
	"net/http"
	"strings"
)

const (
	keyBitSize        = 1024
	verifyTokenLength = 4
)

type Server struct {
	privateKey *rsa.PrivateKey
	publicKey  []byte
	serverID   string
	Skins      map[*mc.Conn]string

	disconnectMessage string
	serverInfoPacket  pk.Packet
}

func NewServer(cfg ServerConfig) (*Server, error) {
	key, err := rsa.GenerateKey(rand.Reader, keyBitSize)
	if err != nil {
		return nil, err
	}

	publicKey, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return nil, err
	}

	server := &Server{
		privateKey: key,
		publicKey:  publicKey,
		serverID:   "",
	}

	return server, server.UpdateConfig(cfg)
}

func (server Server) HandleConn(conn mc.Conn) error {
	packet, err := conn.ReadPacket()
	if err != nil {
		return err
	}

	handshake, err := protocol.ParseHandshakingServerBoundHandshake(packet)
	if err != nil {
		return err
	}

	if handshake.IsStatusRequest() {
		return server.respondToStatusRequest(conn)
	} else if handshake.IsLoginRequest() {
		return server.respondToLoginRequest(conn)
	}

	return nil
}

func (server Server) respondToStatusRequest(conn mc.Conn) error {
	packet, err := conn.ReadPacket()
	if err != nil {
		return err
	}

	if packet.ID != protocol.HandshakingServerBoundRequestPacketID {
		return fmt.Errorf("expexted request protocol \"%d\"; got this %d", protocol.HandshakingServerBoundRequestPacketID, packet.ID)
	}

	if err := conn.WritePacket(server.serverInfoPacket); err != nil {
		return err
	}

	packet, err = conn.ReadPacket()
	if err != nil {
		return err
	}

	if packet.ID != protocol.HandshakingServerBoundPingPacketID {
		return fmt.Errorf("expexted ping protocol id \"%d\"; got this %d", protocol.HandshakingServerBoundPingPacketID, packet.ID)
	}

	return conn.WritePacket(packet)
}

func (server Server) respondToLoginRequest(conn mc.Conn) error {
	packet, err := conn.ReadPacket()
	if err != nil {
		return err
	}

	loginStart, err := protocol.ParseLoginServerBoundLoginStart(packet)
	if err != nil {
		return err
	}

	message := strings.Replace(server.disconnectMessage, "$username", string(loginStart.Name), -1)
	message = fmt.Sprintf("{\"text\":\"%s\"}", message)

	disconnect := protocol.LoginClientBoundDisconnect{
		Reason: pk.Chat(message),
	}

	return conn.WritePacket(disconnect.Marshal())
}

func (server Server) authenticateSession(conn *mc.Conn, username string, sharedSecret []byte) ([16]byte, string, string, error) {
	notchHash := sha1.New()
	notchHash.Update([]byte(server.serverID))
	notchHash.Update(sharedSecret)
	notchHash.Update(server.publicKey)
	hash := notchHash.HexDigest()

	url := fmt.Sprintf(
		"https://sessionserver.mojang.com/session/minecraft/hasJoined?username=%s&serverId=%s",
		username,
		hash,
	)
	resp, err := http.Get(url)
	if err != nil {
		return [16]byte{}, "", "", err
	}

	if resp.StatusCode != http.StatusOK {
		return [16]byte{}, "", "", fmt.Errorf("unable to authenticate session (%s)", resp.Status)
	}

	var profile = struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Properties []struct {
			Name      string `json:"name"`
			Value     string `json:"value"`
			Signature string `json:"signature"`
		} `json:"properties"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return [16]byte{}, "", "", err
	}
	resp.Body.Close()

	id, err := uuid.Parse(profile.ID)
	if err != nil {
		return [16]byte{}, "", "", err
	}

	var skin string
	var signature string
	for _, property := range profile.Properties {
		if property.Name == "textures" {
			skin = property.Value
			signature = property.Signature
			break
		}
	}

	if skin == "" {
		return [16]byte{}, "", "", errors.New("no skin in request")
	}

	return id, skin, signature, nil
}

func (server *Server) SetEncryption(conn *mc.Conn, player *mc.Player) error {
	verifyToken := make([]byte, verifyTokenLength)
	if _, err := rand.Read(verifyToken); err != nil {
		return err
	}

	var encryptionRequest = protocol.LoginClientBoundEncryptionRequest{
		ServerID:    pk.String(server.serverID),
		PublicKey:   pk.ByteArray(server.publicKey),
		VerifyToken: pk.ByteArray(verifyToken),
	}

	if err := conn.WritePacket(encryptionRequest.Marshal()); err != nil {
		return err
	}

	packet, err := conn.ReadPacket()
	if err != nil {
		return err
	}

	encryptionResponse, err := protocol.ParseLoginServerBoundEncryptionResponse(packet)
	if err != nil {
		return err
	}

	respVerifyToken, err := server.privateKey.Decrypt(rand.Reader, encryptionResponse.VerifyToken, nil)
	if err != nil {
		return err
	}

	if bytes.Compare(verifyToken, respVerifyToken) != 0 {
		return errors.New("verify token did not match")
	}

	sharedSecret, err := server.privateKey.Decrypt(rand.Reader, encryptionResponse.SharedSecret, nil)
	if err != nil {
		return err
	}

	id, skin, signature, err := server.authenticateSession(conn, player.Username, sharedSecret)
	if err != nil {
		return err
	}

	player.UUID = id
	player.Skin = skin
	player.SkinSignature = signature

	block, err := aes.NewCipher(sharedSecret)
	if err != nil {
		return err
	}

	conn.SetCipher(
		cfb8.NewEncrypter(block, sharedSecret),
		cfb8.NewDecrypter(block, sharedSecret),
	)

	return nil
}

func (server *Server) SetCustomThreshold(conn, rconn *mc.Conn, threshold int) error {
	setClientThreshold := func() error {
		if err := conn.WritePacket(protocol.LoginClientBoundSetCompression{
			Threshold: pk.VarInt(threshold),
		}.Marshal()); err != nil {
			return err
		}
		conn.Threshold = threshold
		return nil
	}

	packet, err := rconn.ReadPacket()
	if err != nil {
		return err
	}

	switch packet.ID {
	case protocol.LoginClientBoundLoginSuccessPacketID:
		if err := setClientThreshold(); err != nil {
			return err
		}
		return conn.WritePacket(packet)
	case protocol.LoginClientBoundSetCompressionPacketID:
		setCompression, err := protocol.ParseLoginClientBoundSetCompression(packet)
		if err != nil {
			return err
		}
		rconn.Threshold = int(setCompression.Threshold)
		if err := setClientThreshold(); err != nil {
			return err
		}
		return nil
	default:
		return protocol.ErrInvalidPacketID
	}
}

func (server *Server) SetThreshold(conn, rconn *mc.Conn) (int, error) {
	threshold := 0
	packet, err := rconn.ReadPacket()
	if err != nil {
		return threshold, err
	}

	switch packet.ID {
	case protocol.LoginClientBoundLoginSuccessPacketID:
		return threshold, conn.WritePacket(packet)
	case protocol.LoginClientBoundSetCompressionPacketID:
		setCompression, err := protocol.ParseLoginClientBoundSetCompression(packet)
		if err != nil {
			return threshold, err
		}
		threshold = int(setCompression.Threshold)
		rconn.Threshold = threshold
		if err := conn.WritePacket(packet); err != nil {
			return threshold, err
		}
		conn.Threshold = threshold
		return threshold, nil
	default:
		return threshold, protocol.ErrInvalidPacketID
	}
}

func (server *Server) UpdateConfig(cfg ServerConfig) error {
	pingResponse, err := cfg.marshalPingResponse()
	if err != nil {
		return err
	}

	server.serverInfoPacket = protocol.HandshakingServerBoundResponse{
		JSONResponse: pk.String(pingResponse),
	}.Marshal()

	server.disconnectMessage = cfg.DisconnectMessage

	return nil
}

func SniffUsername(conn, rconn mc.Conn) (string, error) {
	// Handshake
	packet, err := conn.ReadPacket()
	if err != nil {
		return "", err
	}

	if err := rconn.WritePacket(packet); err != nil {
		return "", err
	}

	// Login
	packet, err = conn.ReadPacket()
	if err != nil {
		return "", err
	}

	loginStartPacket, err := protocol.ParseLoginServerBoundLoginStart(packet)
	if err != nil {
		return "", err
	}

	if err := rconn.WritePacket(packet); err != nil {
		return "", err
	}

	return string(loginStartPacket.Name), nil
}
