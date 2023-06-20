package srv

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/averageNetAdmin/andproxy/andproto/models"
)

var ServersMutex = &sync.RWMutex{}
var Servers = make(map[int64]*Server)

type Server struct {
	*models.Server
	stats models.ServerStats
	m     sync.RWMutex
	sm    sync.RWMutex
}

func (s *Server) AddConn() {
	s.sm.Lock()
	s.stats.Connections++
	s.sm.Unlock()
}

func (s *Server) RemoveConn() {
	s.sm.Lock()
	s.stats.Connections--
	s.sm.Unlock()
}

func (s *Server) Down() {
	s.sm.Lock()
	s.stats.Down = true
	s.sm.Unlock()
}

func (s *Server) Up() {
	s.sm.Lock()
	s.stats.Down = false
	s.sm.Unlock()
}

func (s *Server) IsDown() bool {
	s.sm.Lock()
	d := s.stats.Down
	s.sm.Unlock()
	return d
}

func (s *Server) Fail() {
	s.sm.Lock()
	s.stats.Fails++
	s.sm.Unlock()
}

func (s *Server) GetFails() int64 {
	s.sm.RLock()
	f := s.stats.Fails
	s.sm.RUnlock()
	return f
}

func (s *Server) DropFails() {
	s.sm.Lock()
	s.stats.Fails = 0
	s.sm.Unlock()
}

func (s *Server) TryConnect() bool {

	var conn net.Conn
	var err error
	s.m.RLock()
	if s.ConnectTimeout != 0 {
		conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", s.Address, s.Port), time.Duration(s.ConnectTimeout))
	} else {
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", s.Address, s.Port))
	}
	s.m.RUnlock()
	s.m.Lock()
	defer s.m.Unlock()
	if err != nil {
		s.Fail()
		if s.GetFails() >= s.MaxFails {
			s.Down()
			s.DropFails()
		}
		return false
	}

	defer conn.Close()

	return true
}
