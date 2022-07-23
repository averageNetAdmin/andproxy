package http

import (
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/averageNetAdmin/andproxy/internal/ranges"
)

type Server struct {
	Addr           string
	DeadLine       time.Duration
	WriteDeadLine  time.Duration
	ReadDeadLine   time.Duration
	Weight         int
	MaxFails       uint64
	MaxConnections int64
	MaxConnectTime time.Duration
	BreakTime      time.Duration
	httpClient     *http.Client

	broken                   bool
	fails                    uint64
	connectionsNumber        uint64
	currentConnectionsNumber int64
}

//	BalanceItem implementation
//
func (s *Server) GetWeight() int {
	return s.Weight
}

//	BalanceItem implementation
//
func (s *Server) GetConnNumber() uint64 {
	return s.connectionsNumber
}

//	Set deadlines and timeout for connections to server
//
func (s *Server) SetTimeout(network, host string) (net.Conn, error) {
	var conn net.Conn
	var err error
	if s.MaxConnectTime != 0 {
		conn, err = net.DialTimeout(network, host, s.MaxConnectTime)
	} else {
		conn, err = net.Dial(network, host)
	}
	if err != nil {
		s.Fail()
		return nil, err
	}
	_ = atomic.AddUint64(&s.connectionsNumber, 1)
	if s.DeadLine != 0 {
		conn.SetDeadline(time.Now().Add(s.DeadLine))
	}
	if s.ReadDeadLine != 0 {
		conn.SetReadDeadline(time.Now().Add(s.ReadDeadLine))
	}
	if s.WriteDeadLine != 0 {
		conn.SetWriteDeadline(time.Now().Add(s.WriteDeadLine))
	}
	return conn, nil
}

func (s *Server) Fail() {
	v := atomic.AddUint64(&s.fails, 1)
	if s.MaxFails%v == 0 {
		s.broken = true
		time.AfterFunc(s.BreakTime, func() {
			s.broken = false
		})
	}
}

func (s *Server) Do(port string, request *http.Request) (*http.Response, error) {
	if atomic.LoadInt64(&s.currentConnectionsNumber) >= s.MaxConnections && s.MaxConnections != 0 {
		return nil, fmt.Errorf("max parallel connections to server reached")
	}
	reqURL := fmt.Sprintf("http://%s:%s%s", s.Addr, port, request.URL.Path)
	req, err := http.NewRequest(request.Method, reqURL, request.Body)
	if err != nil {
		return nil, err
	}
	response, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	atomic.AddInt64(&s.currentConnectionsNumber, -1)
	return response, nil
}

func NewServer(addr string, deadLine, writeDeadLine, readDeadLine, maxConnectTime, breakTime time.Duration,
	weight int, maxFails uint64, maxConnections int64) (*Server, error) {

	if net.ParseIP(addr) == nil {
		return nil, fmt.Errorf("invalid address: %v", addr)
	}
	if weight <= 0 {
		weight = 1
	}
	if maxConnections <= 0 {
		maxConnections = 9223372036854775807
	}
	if breakTime <= 0 {
		breakTime = time.Minute * 2
	}
	if maxFails < 1 {
		maxFails = 5
	}
	srv := &Server{Addr: addr,
		DeadLine:       deadLine,
		WriteDeadLine:  writeDeadLine,
		ReadDeadLine:   readDeadLine,
		MaxConnectTime: maxConnectTime,

		BreakTime:                breakTime,
		Weight:                   weight,
		MaxFails:                 maxFails,
		MaxConnections:           maxConnections,
		broken:                   false,
		fails:                    0,
		connectionsNumber:        0,
		currentConnectionsNumber: 0,
	}

	tr := &http.Transport{
		Dial: srv.SetTimeout,
	}
	cli := &http.Client{
		Transport: tr,
	}

	srv.httpClient = cli
	return srv, nil
}

func ServersFromMap(config map[string]interface{}) ([]*Server, error) {
	var err error
	addr, ok := config["addr"].(string)
	var (
		dl,
		rdl,
		wdl,
		mconntime,
		breaktime time.Duration
		weight   int
		maxfails uint64
		maxconn  int64
	)
	if !ok {
		return nil, fmt.Errorf("invalid server address %v", config["addr"])
	}
	if config["deadline"] != nil {
		dlS, ok := config["deadline"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid server deadline %v", config["deadline"])
		}
		dl, err = time.ParseDuration(dlS)
		if err != nil {
			return nil, err
		}
	}
	if config["readdeadline"] != nil {
		rdlS, ok := config["readdeadline"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid server readdeadline %v", config["readdeadline"])
		}
		rdl, err = time.ParseDuration(rdlS)
		if err != nil {
			return nil, err
		}
	}
	if config["writedeadline"] != nil {
		wdlS, ok := config["writedeadline"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid server writedeadline %v", config["writedeadline"])
		}
		wdl, err = time.ParseDuration(wdlS)
		if err != nil {
			return nil, err
		}
	}
	if config["maxconnectionstime"] != nil {
		mconntimeS, ok := config["maxconnectionstime"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid server maxconnectionstime %v", config["maxconnectionstime"])
		}
		mconntime, err = time.ParseDuration(mconntimeS)
		if err != nil {
			return nil, err
		}
	}
	if config["breaktime"] != nil {
		breaktimeS, ok := config["breaktime"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid server breaktime %v", config["breaktime"])
		}
		breaktime, err = time.ParseDuration(breaktimeS)
		if err != nil {
			return nil, err
		}
	}
	if config["weight"] != nil {
		weight, ok = config["weight"].(int)
		if !ok {
			return nil, fmt.Errorf("invalid server weight %v", config["weight"])
		}
	}
	if config["maxfails"] != nil {
		mf, ok := config["maxfails"].(int)
		if !ok {
			return nil, fmt.Errorf("invalid server maxfails %v", config["maxfails"])
		}
		maxfails = uint64(mf)
	}
	if config["maxconnections"] != nil {
		mc, ok := config["maxconnections"].(int)
		if !ok {
			return nil, fmt.Errorf("invalid server maxconnections %v", config["maxconnections"])
		}
		maxconn = int64(mc)
	}

	addrs, err := ranges.Create(addr)
	if err != nil {
		return nil, err
	}
	srvs := make([]*Server, 0)
	for _, address := range addrs {
		srv, err := NewServer(address, dl, wdl, rdl, mconntime,
			breaktime, weight, maxfails, maxconn)
		if err != nil {
			return nil, err
		}
		srvs = append(srvs, srv)
	}
	return srvs, nil
}
