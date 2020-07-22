package safe

import (
	"github.com/haveachin/infrared/mc"
	"sync"
)

type Players struct {
	sync.RWMutex
	Value map[*mc.Conn]mc.Player
}

func NewPlayers() *Players {
	return &Players{
		RWMutex: sync.RWMutex{},
		Value:   map[*mc.Conn]mc.Player{},
	}
}

func (p *Players) Put(key *mc.Conn, value mc.Player) {
	p.Lock()
	defer p.Unlock()
	p.Value[key] = value
}

func (p *Players) Get(key *mc.Conn) mc.Player {
	p.RLock()
	defer p.RUnlock()
	return p.Value[key]
}

func (p *Players) Remove(key *mc.Conn) {
	p.Lock()
	defer p.Unlock()
	delete(p.Value, key)
}

func (p *Players) Length() int {
	p.RLock()
	defer p.RUnlock()
	return len(p.Value)
}

func (p *Players) Keys() []*mc.Conn {
	p.RLock()
	defer p.RUnlock()

	var conns []*mc.Conn

	for conn := range p.Value {
		conns = append(conns, conn)
	}

	return conns
}

func (p *Players) Values() []mc.Player {
	p.RLock()
	defer p.RUnlock()

	var players []mc.Player

	for _, player := range p.Value {
		players = append(players, player)
	}

	return players
}
