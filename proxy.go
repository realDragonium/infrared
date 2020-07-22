package infrared

import (
	"github.com/haveachin/infrared/callback"
	"github.com/haveachin/infrared/mots"
	"github.com/haveachin/infrared/safe"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net"
	"sync"
	"time"

	"github.com/haveachin/infrared/mc"
	"github.com/haveachin/infrared/mc/protocol"
	"github.com/haveachin/infrared/mc/sim"
	"github.com/haveachin/infrared/process"
)

// Proxy is a TCP server that takes an incoming request and sends it to another
// server, proxying the response back to the client.
type Proxy struct {
	// Modifier modifies traffic that is send between the client and the server
	Modifier []mots.InterceptFunc

	domainName    safe.String
	listenTo      safe.String
	proxyTo       safe.String
	timeout       safe.Duration
	cancelTimeout func()
	Players       safe.Players

	server    safe.Server
	Process   safe.Process
	logWriter *callback.LogWriter

	logger        zerolog.Logger
	loggerOutputs []io.Writer
}

// NewProxy takes a config an creates a new proxy based on it
func NewProxy(cfg ProxyConfig) (*Proxy, error) {
	logWriter := &callback.LogWriter{}

	server, err := sim.NewServer(cfg.Server)
	if err != nil {
		return nil, err
	}

	proxy := Proxy{
		Modifier:      []mots.InterceptFunc{},
		Players:       *safe.NewPlayers(),
		cancelTimeout: nil,
		server:        safe.Server{Value: *server},
		logWriter:     logWriter,
		loggerOutputs: []io.Writer{logWriter},
	}

	if err := proxy.updateConfig(cfg); err != nil {
		return nil, err
	}

	proxy.overrideLogger(log.Logger)

	return &proxy, nil
}

func (proxy *Proxy) AddLoggerOutput(w io.Writer) {
	proxy.loggerOutputs = append(proxy.loggerOutputs, w)
	proxy.logger = proxy.logger.Output(io.MultiWriter(proxy.loggerOutputs...))
}

func (proxy *Proxy) overrideLogger(logger zerolog.Logger) zerolog.Logger {
	proxy.logger = logger.With().
		Str("destinationAddress", proxy.proxyTo.Read()).
		Str("domainName", proxy.domainName.Read()).Logger().
		Output(io.MultiWriter(proxy.loggerOutputs...))

	return proxy.logger
}

// HandleConn takes a minecraft client connection and it's initial handshake packet
// and relays all following packets to the remote connection (proxyTo)
func (proxy *Proxy) HandleConn(conn mc.Conn) error {
	connAddr := conn.RemoteAddr().String()
	logger := proxy.logger.With().Str("connection", connAddr).Logger()
	state := safe.State{Value: protocol.StateHandshaking}

	packet, err := conn.PeekPacket()
	if err != nil {
		return err
	}

	handshake, err := protocol.ParseHandshakingServerBoundHandshake(packet)
	if err != nil {
		return err
	}

	rconn, err := mc.DialTimeout(proxy.proxyTo.Read(), time.Millisecond*500)
	if err != nil {
		defer conn.Close()
		if handshake.IsStatusRequest() {
			return proxy.server.HandleConn(conn)
		}

		isProcessRunning, err := proxy.Process.IsRunning()
		if err != nil {
			logger.Err(err).Interface(callback.EventKey, callback.ErrorEvent).Msg("Could not determine if the container is running")
			return proxy.server.HandleConn(conn)
		}

		if isProcessRunning {
			return proxy.server.HandleConn(conn)
		}

		logger.Info().Interface(callback.EventKey, callback.ContainerStartEvent).Msg("Starting container")
		if err := proxy.Process.Start(); err != nil {
			logger.Err(err).Interface(callback.EventKey, callback.ErrorEvent).Msg("Could not start the container")
			return proxy.server.HandleConn(conn)
		}

		proxy.startTimeout()

		return proxy.server.HandleConn(conn)
	}

	if handshake.IsLoginRequest() {
		state.Value = protocol.StateLogin

		username, err := sim.SniffUsername(conn, rconn)
		if err != nil {
			return err
		}

		proxy.stopTimeout()
		proxy.Players.Put(&conn, mc.Player{Username: username})
		logger = logger.With().Str("username", username).Logger()
		logger.Info().Interface(callback.EventKey, callback.PlayerJoinEvent).Msgf("%s joined the game", username)

		defer func() {
			logger.Info().Interface(callback.EventKey, callback.PlayerLeaveEvent).Msgf("%s left the game", username)
			proxy.Players.Remove(&conn)

			if proxy.Players.Length() <= 0 {
				proxy.startTimeout()
			}
		}()

		player := proxy.Players.Get(&conn)
		if err := proxy.server.SetEncryption(&conn, &player); err != nil {
			return err
		}

		logger.Debug().Msg("Encryption successful")

		//threshold := 256
		threshold, err := proxy.server.SetThreshold(&conn, &rconn)
		if err != nil {
			return err
		}

		logger.Debug().Msgf("Threshold set to %d", threshold)

		// Send LoginSuccess
		packet, err = rconn.ReadPacket()
		if err != nil {
			return err
		}

		loginSuccess, err := protocol.ParseServerLoginLoginSuccess(packet)
		if err != nil {
			return err
		}

		player.OfflineUUID = loginSuccess.UUID
		proxy.Players.Put(&conn, player)
		loginSuccess.UUID = player.UUID

		if err := conn.WritePacket(loginSuccess.Marshal()); err != nil {
			return err
		}

		state.Value = protocol.StatePlay
	}

	wg := sync.WaitGroup{}

	var pipe = func(src, dst *mc.Conn, modifiers ...mots.InterceptFunc) {
		defer wg.Done()

		author := mots.AuthorServer
		if src == &conn {
			author = mots.AuthorClient
		}

		for {
			packet, err := src.ReadPacket()
			if err != nil {
				return
			}

			msg := mots.Message{
				State:  protocol.StatePlay,
				Author: author,
				Packet: packet,
				Src:    src,
				Dst:    dst,
				Cancel: false,
			}

			for _, modifier := range modifiers {
				modifier(&msg)
			}

			if msg.Cancel {
				continue
			}

			if err := dst.WritePacket(msg.Packet); err != nil {
				return
			}
		}
	}

	var modifiers = []mots.InterceptFunc{
		mots.PlayerInfoSkinMiddleware(&proxy.Players),
		mots.SpawnPlayerSkinMiddleware(&proxy.Players),
		mots.ChatMessageCommandMiddleware,
	}

	wg.Add(2)
	go pipe(&conn, &rconn, modifiers...)
	go pipe(&rconn, &conn, modifiers...)
	wg.Wait()

	conn.Close()
	rconn.Close()

	return nil
}

// updateConfig is a callback function that handles config changes
func (proxy *Proxy) updateConfig(cfg ProxyConfig) error {
	if cfg.ProxyTo == "" {
		ip, err := net.ResolveIPAddr(cfg.Process.Docker.DNSServer, cfg.Process.Docker.ContainerName)
		if err != nil {
			return err
		}

		cfg.ProxyTo = ip.String()
	}

	timeout, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return err
	}

	proc, err := process.New(cfg.Process)
	if err != nil {
		return err
	}

	server := proxy.server.Read()
	if err := server.UpdateConfig(cfg.Server); err != nil {
		return err
	}

	logWriter, err := callback.NewLogWriter(cfg.CallbackLog)
	if err != nil {
		return err
	}

	proxy.logWriter.URL = logWriter.URL
	proxy.logWriter.Events = logWriter.Events

	proxy.domainName.Write(cfg.DomainName)
	proxy.listenTo.Write(cfg.ListenTo)
	proxy.proxyTo.Write(cfg.ProxyTo)
	proxy.timeout.Write(timeout)
	proxy.Process.Update(proc)
	proxy.server.Update(server)

	return nil
}

func (proxy *Proxy) startTimeout() {
	if proxy.cancelTimeout != nil {
		proxy.stopTimeout()
	}

	timer := time.AfterFunc(proxy.timeout.Read(), func() {
		proxy.logger.Info().
			Interface(callback.EventKey, callback.ContainerStopEvent).
			Msgf("Stopping container")
		if err := proxy.Process.Stop(); err != nil {
			proxy.logger.Err(err).
				Interface(callback.EventKey, callback.ErrorEvent).
				Msg("Failed to stop the container")
		}
	})

	proxy.cancelTimeout = func() {
		timer.Stop()
		proxy.logger.Debug().Msg("Timeout canceled")
	}

	proxy.logger.Info().
		Interface(callback.EventKey, callback.ContainerTimeoutEvent).
		Msgf("Timing out in %s", proxy.timeout.Read())
}

func (proxy *Proxy) stopTimeout() {
	if proxy.cancelTimeout == nil {
		return
	}

	proxy.cancelTimeout()
	proxy.cancelTimeout = nil
}

func (proxy *Proxy) Close() {
	for _, conn := range proxy.Players.Keys() {
		if err := conn.Close(); err != nil {
			proxy.logger.Err(err)
		}
	}
}
