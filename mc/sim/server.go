package sim

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/haveachin/infrared/mc/cfb8"
	"strings"

	"github.com/haveachin/infrared/mc"
	pk "github.com/haveachin/infrared/mc/packet"
	"github.com/haveachin/infrared/mc/protocol"
)

const (
	keyBitSize        = 1024
	verifyTokenLength = 4
)

type Server struct {
	privateKey *rsa.PrivateKey
	publicKey  []byte

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
	}

	return server, server.UpdateConfig(cfg)
}

func (server Server) HandleConn(conn mc.Conn) error {
	packet, err := conn.ReadPacket()
	if err != nil {
		return err
	}

	handshake, err := protocol.ParseSLPHandshake(packet)
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

	if packet.ID != protocol.SLPRequestPacketID {
		return fmt.Errorf("expexted request protocol \"%d\"; got this %d", protocol.SLPRequestPacketID, packet.ID)
	}

	if err := conn.WritePacket(server.serverInfoPacket); err != nil {
		return err
	}

	packet, err = conn.ReadPacket()
	if err != nil {
		return err
	}

	if packet.ID != protocol.SLPPingPacketID {
		return fmt.Errorf("expexted ping protocol id \"%d\"; got this %d", protocol.SLPPingPacketID, packet.ID)
	}

	return conn.WritePacket(packet)
}

func (server Server) respondToLoginRequest(conn mc.Conn) error {
	packet, err := conn.ReadPacket()
	if err != nil {
		return err
	}

	loginStart, err := protocol.ParseClientLoginStart(packet)
	if err != nil {
		return err
	}

	message := strings.Replace(server.disconnectMessage, "$username", string(loginStart.Name), -1)
	message = fmt.Sprintf("{\"text\":\"%s\"}", message)

	disconnect := protocol.ServerLoginDisconnect{
		Reason: pk.Chat(message),
	}

	return conn.WritePacket(disconnect.Marshal())
}

func (server *Server) SetEncryption(conn *mc.Conn) error {
	verifyToken := make([]byte, verifyTokenLength)
	if _, err := rand.Read(verifyToken); err != nil {
		return err
	}

	var encryptionRequest = protocol.ServerLoginEncryptionRequest{
		ServerID:    pk.String(""),
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

	encryptionResponse, err := protocol.ParseClientLoginEncryptionResponse(packet)
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
		if err := conn.WritePacket(protocol.ServerLoginSetCompression{
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
	case protocol.ServerLoginLoginSuccessPacketID:
		if err := setClientThreshold(); err != nil {
			return err
		}
		return conn.WritePacket(packet)
	case protocol.ServerLoginSetCompressionPacketID:
		setCompression, err := protocol.ParseServerLoginSetCompression(packet)
		if err != nil {
			return err
		}
		rconn.Threshold = int(setCompression.Threshold)
		if err := setClientThreshold(); err != nil {
			return err
		}
		// Send LoginSuccess
		packet, err = rconn.ReadPacket()
		if err != nil {
			return err
		}
		return conn.WritePacket(packet)
	default:
		return protocol.ErrInvalidPacketID
	}
}

func (server *Server) SetThreshold(conn, rconn *mc.Conn) (int, error) {
	threshold := -1
	packet, err := rconn.ReadPacket()
	if err != nil {
		return threshold, err
	}

	switch packet.ID {
	case protocol.ServerLoginLoginSuccessPacketID:
		return threshold, conn.WritePacket(packet)
	case protocol.ServerLoginSetCompressionPacketID:
		setCompression, err := protocol.ParseServerLoginSetCompression(packet)
		if err != nil {
			return threshold, err
		}
		threshold = int(setCompression.Threshold)
		rconn.Threshold = threshold
		if err := conn.WritePacket(packet); err != nil {
			return threshold, err
		}
		conn.Threshold = threshold
		// Send LoginSuccess
		packet, err = rconn.ReadPacket()
		if err != nil {
			return threshold, err
		}
		return threshold, conn.WritePacket(packet)
	default:
		return threshold, protocol.ErrInvalidPacketID
	}
}

func (server *Server) UpdateConfig(cfg ServerConfig) error {
	pingResponse, err := cfg.marshalPingResponse()
	if err != nil {
		return err
	}

	server.serverInfoPacket = protocol.SLPResponse{
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

	loginStartPacket, err := protocol.ParseClientLoginStart(packet)
	if err != nil {
		return "", err
	}

	if err := rconn.WritePacket(packet); err != nil {
		return "", err
	}

	return string(loginStartPacket.Name), nil
}
