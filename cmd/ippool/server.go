package ippool

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//
//	contains server data exchange statistics
//
type ServerStats struct {
	DataSended          uint64
	DataReceived        uint64
	AvgDataExchangeTime uint64
	AvgConnectTime      uint64
	ConnectionsNumber   uint64
}

//
//	reset statistics to zero
//
func (s *ServerStats) Reset() {
	s.DataSended = 0
	s.DataReceived = 0
	s.AvgDataExchangeTime = 0
	s.AvgConnectTime = 0
	s.ConnectionsNumber = 0
}

//
//	representation of the server
//
type Server struct {
	Addr                     string
	Broken                   bool
	Weight                   int
	Priority                 float64
	Fails                    int
	MaxFails                 int
	stats                    *ServerStats
	CurrentConnectionsNumber int
	BreakTime                int
	logger                   *log.Logger
}

//
//	create and return new server. recive server address or domain name, server weight, maximal fail number,
//	and break time after maximal allowed fails
//
func NewServer(addr string, weight, maxFails, breakTime int) *Server {
	if weight < 1 {
		weight = 1
	}
	if maxFails < 1 {
		maxFails = 1
	}
	if breakTime < 0 {
		breakTime = 0
	}
	srvstats := new(ServerStats)

	return &Server{
		Addr:                     addr,
		Weight:                   weight,
		MaxFails:                 maxFails,
		Broken:                   false,
		Priority:                 1,
		Fails:                    0,
		stats:                    srvstats,
		CurrentConnectionsNumber: 0,
		BreakTime:                breakTime,
	}
}

//
//	connct to the server, recieve protokol and port, return conncetion (net.Conn) or error
//
func (s *Server) Connect(proto, port string) (net.Conn, error) {
	timerStart := time.Now()
	conn, err := net.Dial(proto, net.JoinHostPort(s.Addr, port))
	if err != nil {
		s.Fail()
		return nil, err
	}
	dur := time.Since(timerStart)
	duration := dur.Microseconds()
	s.stats.ConnectionsNumber++
	s.stats.AvgConnectTime = (s.stats.AvgConnectTime*s.stats.ConnectionsNumber - 1/
		s.stats.ConnectionsNumber + uint64(duration)) / s.stats.ConnectionsNumber
	s.CurrentConnectionsNumber++
	s.logger.Printf("successfull connect to server. Connect time: %v\n", dur.String())
	return conn, nil
}

//
//	create a data pipe between client and server connections. end when ends connection session
//
func (s *Server) ExchangeData(client net.Conn, server net.Conn) {
	start := time.Now()
	var dataRecieved int64
	var dataSended int64
	var err error
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		dataRecieved, err = io.Copy(client, server)
		wg.Done()
	}()
	go func() {
		dataSended, err = io.Copy(server, client)
		wg.Done()
	}()
	if err != nil {
		s.logger.Println(err)
	}
	wg.Wait()
	dur := time.Since(start)
	duration := dur.Microseconds()
	s.stats.AvgDataExchangeTime = (s.stats.AvgDataExchangeTime*s.stats.ConnectionsNumber - 1/
		s.stats.ConnectionsNumber + uint64(duration)) / s.stats.ConnectionsNumber
	s.logger.Printf("successfull data exchange with %v. %d, bytes received, %d bytes sended, session time: %v\n",
		client.RemoteAddr().String(), dataRecieved, dataSended, dur.String())
}

//
//	close connection
//
func (s *Server) Disconnect(conn net.Conn) error {
	err := conn.Close()
	if err != nil {
		return err
	}
	s.logger.Printf("connection closed\n")
	s.CurrentConnectionsNumber--
	return nil
}

//
//	increment server fails quantity. if server fails == max fails, server become broken
//
func (s *Server) Fail() {
	s.Fails++
	s.logger.Printf("server is failed %d times\n", s.Fails)
	if s.MaxFails%s.Fails == 0 {
		s.Broken = true
		s.logger.Printf("max fails reached, server is break\n")
		time.AfterFunc(time.Second*time.Duration(s.BreakTime), func() {
			s.logger.Printf("server is return to work\n")
			s.Broken = false
		})
	}
}

//
//	setting server log file
//
func (s *Server) SetLogFile(logDir string) error {
	s.Priority = 100
	err := os.MkdirAll(logDir, 0644)
	if err != nil {
		return err
	}
	logFile := fmt.Sprintf("%s/%s.log", logDir, s.Addr)
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	logger := log.New(file, fmt.Sprintf("server %s ", s.Addr), log.LstdFlags)
	logger.SetFlags(log.LstdFlags)
	s.logger = logger
	return nil
}

//
//	set new server config
//
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
		s.BreakTime = w
	} else if strings.HasPrefix(breakTime, "*") {
		b, err := strconv.Atoi(breakTime[1:])
		if err != nil {
			return err
		}
		s.BreakTime *= b
	} else {
		b, err := strconv.Atoi(breakTime)
		if err != nil {
			return err
		}
		if s.BreakTime == 1 {
			s.BreakTime = b
		}
	}

	return nil
}
