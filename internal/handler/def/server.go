package def

import (
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/averageNetAdmin/andproxy/internal/ranges"
)

//	Representaion of server - everything that have ip address and can get requests
//
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

	broken                   bool
	fails                    uint64
	connectionsNumber        uint64
	currentConnectionsNumber int64
}

//	Getter to match BalanceItem interface
//
func (s *Server) GetWeight() int {
	return s.Weight
}

//	Getter to match BalanceItem interface
//
func (s *Server) GetConnNumber() uint64 {
	return s.connectionsNumber
}

//	Increment server fail number
//	If fail number reach MaxFails server sleep time equal BreakTime
//
func (s *Server) Fail() {
	v := atomic.AddUint64(&s.fails, 1)
	if s.MaxFails%v == 0 {
		s.broken = true
		time.AfterFunc(s.BreakTime, func() {
			s.broken = false
		})
	}
}

//	Connect to server
//
//
func (s *Server) Connect(proto string, port string) (net.Conn, error) {
	if atomic.LoadInt64(&s.currentConnectionsNumber) == s.MaxConnections {
		return nil, fmt.Errorf("max parallel connections to server reached")
	}
	var conn net.Conn
	var err error
	if s.MaxConnectTime != 0 {
		conn, err = net.DialTimeout(proto, net.JoinHostPort(s.Addr, port), s.MaxConnectTime)
	} else {
		conn, err = net.Dial(proto, net.JoinHostPort(s.Addr, port))
	}
	if err != nil {
		s.Fail()
		return nil, err
	}
	atomic.AddInt64(&s.currentConnectionsNumber, 1)

	return conn, nil
}

//	Make pipe between client connection and server connection
//	Can have deadlines
//
func (s *Server) Exchange(client net.Conn, server net.Conn) {
	start := time.Now()
	if s.DeadLine != 0 {
		server.SetDeadline(start.Add(s.DeadLine))
	}
	if s.ReadDeadLine != 0 {
		server.SetReadDeadline(start.Add(s.ReadDeadLine))
	}
	if s.WriteDeadLine != 0 {
		server.SetWriteDeadline(start.Add(s.WriteDeadLine))
	}
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		_, _ = io.Copy(client, server)
		wg.Done()
	}()
	go func() {
		_, _ = io.Copy(server, client)
		wg.Done()
	}()

	wg.Wait()
	atomic.AddInt64(&s.currentConnectionsNumber, -1)
	client.Close()
	server.Close()
}

//
//
func NewServer(addr string, deadLine, writeDeadLine, readDeadLine, maxConnectTime, breakTime time.Duration,
	weight int, maxFails uint64, maxConnections int64, logDir string) (*Server, error) {

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

	return &Server{Addr: addr,
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
	}, nil
}

//	Create server objects from config map
//
func ServersFromMap(config map[string]interface{}, logDir string) ([]*Server, error) {
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
		maxfails, ok = config["maxfails"].(uint64)
		if !ok {
			return nil, fmt.Errorf("invalid server maxfails %v", config["maxfails"])
		}
	}
	if config["maxconnections"] != nil {
		maxconn, ok = config["maxconnections"].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid server maxconnections %v", config["maxconnections"])
		}
	}

	addrs, err := ranges.Create(addr)
	if err != nil {
		return nil, err
	}
	srvs := make([]*Server, 0)
	for _, address := range addrs {
		srv, err := NewServer(address, dl, wdl, rdl, mconntime,
			breaktime, weight, maxfails, maxconn, logDir)
		if err != nil {
			return nil, err
		}
		srvs = append(srvs, srv)
	}
	return srvs, nil
}
