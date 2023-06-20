package l4handler

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/averageNetAdmin/andproxy/andproto/models"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/srv"
	"github.com/pkg/errors"
)

var L4HandlersMutex = &sync.RWMutex{}
var L4Handlers = make(map[int64]*L4Handler)

type L4Handler struct {
	*models.L4Handler
	stats models.HandlerStats
	m     sync.RWMutex
	sm    sync.RWMutex
}

func (h *L4Handler) AddConn() {
	h.sm.Lock()
	h.stats.Connections++
	h.sm.Unlock()
}

func (h *L4Handler) RemoveConn() {
	h.sm.Lock()
	h.stats.Connections--
	h.sm.Unlock()
}

// Run handler job gorutine
func (s *L4Handler) StartListen() {
	go s.listen()
}

// Get requests conn and delegate they to handle function
func (s *L4Handler) listen() {

	funcName := "listen"

	listener, err := net.Listen(s.Protocol, fmt.Sprintf("0.0.0.0:%d", s.Port))
	if err != nil {
		err = errors.WithMessage(err, "net.Listen()")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			err = errors.WithMessage(err, "listener.Accept()")
			config.Glogger.WithField("func", funcName).Error(err)
			return
		}
		go s.handle(conn)
	}

}

func (h *L4Handler) handle(clientConn net.Conn) {

	funcName := "handle"

	defer clientConn.Close()

	clinetIp, _, err := net.SplitHostPort(clientConn.RemoteAddr().String())
	if err != nil {
		err = errors.WithMessage(err, "net.SplitHostPort()")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}

	h.m.RLock()
	defer h.m.RUnlock()
	pind := int64(0)
	srv.ProxyFiltersMutex.RLock()
	for _, find := range h.ProxyFilters {
		f, exists := srv.ProxyFilters[find]

		if !exists {
			err = errors.New("access denied")
			config.Glogger.WithField("func", funcName).Error(err)
			return
		}
		pind, err = f.GetPool(clinetIp)
		if err == nil {
			break
		}
	}
	srv.ProxyFiltersMutex.RUnlock()
	if err != nil {
		err = errors.WithMessage(err, "srv.ProxyFiltersMutex()")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}

	srv.ServerPoolsMutex.RLock()
	p, exists := srv.ServerPools[pind]
	srv.ServerPoolsMutex.RUnlock()
	if !exists {
		err = errors.New("access denied")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}
	sind := p.GetSrv(clinetIp)

	srv.ServersMutex.RLock()
	s, exists := srv.Servers[sind]
	srv.ServersMutex.RUnlock()
	if !exists {
		err = errors.New("access denied")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}

	//	Check and set deadlines
	//

	start := time.Now()
	if h.Deadline != 0 {
		clientConn.SetDeadline(start.Add(time.Duration(h.Deadline)))
	}
	if h.ReadDeadline != 0 {
		clientConn.SetReadDeadline(start.Add(time.Duration(h.ReadDeadline)))
	}
	if h.WriteDeadline != 0 {
		clientConn.SetWriteDeadline(start.Add(time.Duration(h.WriteDeadline)))
	}

	// if atomic.LoadInt64(&s.currentConnectionsNumber) == s.MaxConnections {
	// 	return nil, fmt.Errorf("max parallel connections to server reached")
	// }

	connSocket := fmt.Sprintf("%s:%d", s.Address, s.Port)
	var srvConn net.Conn
	if s.ConnectTimeout != 0 {
		srvConn, err = net.DialTimeout(h.Protocol, connSocket, time.Duration(s.ConnectTimeout))
	} else {
		srvConn, err = net.Dial(h.Protocol, connSocket)
	}
	if err != nil {
		err = errors.WithMessage(err, "net.Dial()")
		config.Glogger.WithField("func", funcName).Error(err)
		return
	}
	defer srvConn.Close()

	s.AddConn()
	h.AddConn()
	defer func() {
		s.RemoveConn()
		h.RemoveConn()
	}()

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		_, _ = io.Copy(clientConn, srvConn)
		wg.Done()
	}()
	go func() {
		_, _ = io.Copy(srvConn, clientConn)
		wg.Done()
	}()

	wg.Wait()

}
