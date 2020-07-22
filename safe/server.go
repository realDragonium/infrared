package safe

import (
	"github.com/haveachin/infrared/mc"
	"github.com/haveachin/infrared/mc/sim"
	"sync"
)

type Server struct {
	sync.Mutex
	Value sim.Server
}

func (safe *Server) Read() sim.Server {
	safe.Lock()
	defer safe.Unlock()
	return safe.Value
}

func (safe *Server) Update(value sim.Server) {
	safe.Lock()
	defer safe.Unlock()
	safe.Value = value
}

func (safe *Server) HandleConn(conn mc.Conn) error {
	safe.Lock()
	defer safe.Unlock()
	return safe.Value.HandleConn(conn)
}

func (safe *Server) SetCustomThreshold(conn, rconn *mc.Conn, threshold int) error {
	safe.Lock()
	defer safe.Unlock()
	return safe.Value.SetCustomThreshold(conn, rconn, threshold)
}

func (safe *Server) SetThreshold(conn, rconn *mc.Conn) (int, error) {
	safe.Lock()
	defer safe.Unlock()
	return safe.Value.SetThreshold(conn, rconn)
}

func (safe *Server) SetEncryption(conn *mc.Conn, player *mc.Player) error {
	safe.Lock()
	defer safe.Unlock()
	return safe.Value.SetEncryption(conn, player)
}
