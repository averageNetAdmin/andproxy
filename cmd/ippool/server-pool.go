package ippool

import (
	"fmt"
	"log"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	Addr                     netip.Addr
	Broken                   bool
	Weight                   int
	Priority                 int
	fails                    int
	MaxFails                 int
	AvgDataExchangeTime      int64
	AvgConnectTime           int64
	ConnectionsNumber        int
	CurrentConnectionsNumber int
	breakTime                int
	logger                   *log.Logger
}

func NewServer(addr netip.Addr, weight, maxFails, breakTime int) *Server {
	if weight < 1 {
		weight = 1
	}
	if maxFails < 1 {
		maxFails = 1
	}
	if breakTime < 0 {
		breakTime = 0
	}

	return &Server{
		Addr:                     addr,
		Weight:                   weight,
		MaxFails:                 maxFails,
		Broken:                   false,
		Priority:                 1,
		fails:                    0,
		AvgDataExchangeTime:      0,
		AvgConnectTime:           0,
		ConnectionsNumber:        0,
		CurrentConnectionsNumber: 0,
		breakTime:                breakTime,
	}
}

func (s *Server) Connect(proto, port string) (net.Conn, error) {
	timerStart := time.Now()
	conn, err := net.Dial(proto, s.Addr.String()+":"+port)
	if err != nil {
		s.Fail()
		return nil, err
	}
	dur := time.Since(timerStart)
	duration := dur.Microseconds()
	s.ConnectionsNumber++
	if s.ConnectionsNumber >= 2000000000 {
		s.ConnectionsNumber = 10
	}
	s.AvgConnectTime = (s.AvgConnectTime*int64(s.ConnectionsNumber-1)/
		int64(s.ConnectionsNumber) + duration) / int64(s.ConnectionsNumber)
	s.CurrentConnectionsNumber++
	s.logger.Printf("successfull connect to server. Connect time: %v\n", dur.String())
	return conn, nil
}

func (s *Server) ExchangeData(client net.Conn, server net.Conn) {
	start := time.Now()
	dataRecieved := 0
	dataSended := 0
	packetsReceived := 0
	packetsSended := 0
	for {
		packet := make([]byte, 1500)
		n, err := client.Read(packet)
		if err != nil {
			break
		}
		dataRecieved += n
		packetsReceived++
		_, err = server.Write(packet[:n])
		if err != nil {
			s.logger.Println(err)
			return
		}
		n, err = server.Read(packet)
		if err != nil {
			break
		}
		n, err = client.Write(packet[:n])
		dataSended += n
		packetsSended++
		if err != nil {
			s.logger.Println(err)
			return
		}
		fmt.Println("qwer")
	}
	fmt.Println("asdf")
	dur := time.Since(start)
	duration := dur.Microseconds()
	s.AvgDataExchangeTime = (s.AvgDataExchangeTime*int64(s.ConnectionsNumber-1)/
		int64(s.ConnectionsNumber) + duration) / int64(s.ConnectionsNumber)
	s.logger.Printf("successfull data exchange with %v. %d packets recieved %d packets sended, %d, bytes received, %d bytes sended, session time: %v\n",
		client.RemoteAddr().String(), packetsReceived, packetsSended, dataRecieved, dataSended, dur.String())
}

func (s *Server) Disconnect(conn net.Conn) error {
	err := conn.Close()
	if err != nil {
		return err
	}
	s.logger.Printf("connection closed\n")
	s.CurrentConnectionsNumber--
	fmt.Printf("Connect closed")
	return nil
}

func (s *Server) Fail() {
	s.fails++
	s.logger.Printf("server is failed %d times\n", s.fails)
	if s.fails == s.MaxFails {
		s.Broken = true
		s.fails = 0
		s.logger.Printf("max fails reached, server is break\n")
		time.AfterFunc(time.Second*time.Duration(s.breakTime), func() {
			s.logger.Printf("server is return to work\n")
			s.Broken = false
		})
	}
}

func (s *Server) SetLogFile(logDir string) error {
	s.Priority = 100
	err := os.MkdirAll(logDir, 0644)
	if err != nil {
		return err
	}
	logFile := logDir + "/" + s.Addr.String() + ".log"
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logger := log.New(file, "andproxy server log: ", log.LstdFlags)
	logger.SetFlags(log.LstdFlags)
	s.logger = logger
	return nil
}

func (s *Server) CheckLogFile() {
	fmt.Println(s.logger)
}

//
//	the struct that keep servers addresses
//	contain only IP addresses
//
type ServerPool struct {
	Servers []Server
}

func (s *ServerPool) SetLogFile(logDir string) error {
	for i := 0; i < len(s.Servers); i++ {
		err := s.Servers[i].SetLogFile(logDir)
		if err != nil {
			return err
		}
	}
	return nil
}

//
//	check the server IP in pool or not
//	return true if pool contain IP
//
func (p *ServerPool) Contains(searchIP string) (bool, error) {
	ip, err := netip.ParseAddr(searchIP)
	if err != nil {
		return false, err
	}
	for _, srv := range p.Servers {
		if srv.Addr.Compare(ip) == 0 {
			return true, nil
		}
	}
	return false, nil
}

//
//	add IP address in servers pool
//
func (p *ServerPool) Add(addresses string, config map[string]interface{}) error {

	var (
		weight    int
		maxFails  int
		breakTime int
	)
	if val, ok := config["weight"]; ok && val != nil {
		weight, ok = val.(int)
		if !ok {
			return fmt.Errorf("error: invalid weight value: %v", config)
		}
	} else {
		weight = 1
	}
	if val, ok := config["maxFails"]; ok && val != nil {
		maxFails, ok = val.(int)
		if !ok {
			return fmt.Errorf("error: invalid maxFails value: %v", config)
		}
	} else {
		maxFails = 1
	}
	if val, ok := config["breakTime"]; ok && val != nil {
		breakTime, ok = val.(int)
		if !ok {
			return fmt.Errorf("error: invalid breakTime value: %v", config)
		}
	} else {
		breakTime = 0
	}
	isRange := strings.Contains(addresses, "-")
	if isRange {
		rng, err := CreateIPRange(addresses)
		if err != nil {
			return err
		}
		for _, addr := range rng {
			srv := NewServer(addr, weight, maxFails, breakTime)
			p.Servers = append(p.Servers, *srv)
		}
	} else {
		addr, err := netip.ParseAddr(addresses)
		if err != nil {
			return err
		}
		srv := NewServer(addr, weight, maxFails, breakTime)
		p.Servers = append(p.Servers, *srv)
	}
	return nil
}

//
//	add IP address in servers pool from array
//
func (p *ServerPool) AddFromMap(m map[string]interface{}) error {
	for addresses, config := range m {
		if config == nil {
			err := p.Add(addresses, make(map[string]interface{}))
			if err != nil {
				return err
			}
			continue
		}
		conf, ok := config.(map[string]interface{})
		if !ok {
			return fmt.Errorf("syntax error: invalid server config syntax in %v", config)
		}
		err := p.Add(addresses, conf)
		if err != nil {
			return err
		}
	}
	return nil
}

//
//	create and return new ServerPool
//
func NewServerPool(m map[string]interface{}) (*ServerPool, error) {
	p := new(ServerPool)
	err := p.AddFromMap(m)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *ServerPool) SetConfig(weight, maxFails, breakTime string) error {
	for i := 0; i < len(p.Servers); i++ {
		err := p.Servers[i].SetConfig(weight, maxFails, breakTime)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) SetConfig(weight, maxFails, breakTime string) error {

	if strings.HasPrefix(weight, "!") {
		w, err := strconv.Atoi(weight[1:])
		if err != nil {
			return err
		}
		s.Weight = w
	} else if strings.HasPrefix(weight, "*") {
		w, err := strconv.Atoi(weight[1:])
		if err != nil {
			return err
		}
		s.Weight *= w
	} else {
		w, err := strconv.Atoi(weight)
		if err != nil {
			return err
		}
		if s.Weight == 1 {
			s.Weight = w
		}
	}

	if strings.HasPrefix(maxFails, "!") {
		f, err := strconv.Atoi(maxFails[1:])
		if err != nil {
			return err
		}
		s.MaxFails = f
	} else if strings.HasPrefix(maxFails, "*") {
		f, err := strconv.Atoi(maxFails[1:])
		if err != nil {
			return err
		}
		s.MaxFails *= f
	} else {
		f, err := strconv.Atoi(maxFails)
		if err != nil {
			return err
		}
		if s.MaxFails == 1 {
			s.MaxFails = f
		}
	}

	if strings.HasPrefix(breakTime, "!") {
		w, err := strconv.Atoi(breakTime[1:])
		if err != nil {
			return err
		}
		s.breakTime = w
	} else if strings.HasPrefix(breakTime, "*") {
		b, err := strconv.Atoi(breakTime[1:])
		if err != nil {
			return err
		}
		s.breakTime *= b
	} else {
		b, err := strconv.Atoi(breakTime)
		if err != nil {
			return err
		}
		if s.breakTime == 1 {
			s.breakTime = b
		}
	}

	return nil
}
