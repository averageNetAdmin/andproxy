package def

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/averageNetAdmin/andproxy/internal/client"
	"gopkg.in/yaml.v3"
)

// Contain all info about tcp and udp handlers
//
type Handler struct {
	Protocol                 string
	Port                     string
	connectionsNumber        uint64
	currentconnectionsNumber int64
	rejected                 uint64
	logger                   *log.Logger

	Accept         *client.Sources
	Deny           *client.Sources
	Servers        *Pool
	IPFilter       []*IPFilter
	Toport         int
	LogPath        string
	DeadLine       time.Duration
	WriteDeadLine  time.Duration
	ReadDeadLine   time.Duration
	MaxConnectTime time.Duration
	MaxConnections int64
	OverFlow       string
}

//	Create new handler from yaml file
//
func NewHandler(configPath, protocol, port string) (*Handler, error) {
	// read config file
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	
	// parse config to map
	config := make(map[string]interface{}, 0)
	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		return nil, err
	}

	// check is log dir checked manually
	logDir, ok := config["logdir"].(string)
	// if not create log in defult dir
	if !ok {
		logDir = fmt.Sprintf("/var/log/andproxy/%s_%s/", protocol, port)
	}
	err = os.MkdirAll(logDir, 0644)
	if err != nil {
		return nil, err
	}

	// parse accepted clients address if field not empty 
	accept, err := client.New()
	if err != nil {
		return nil, err
	}
	acceptArr, ok := config["accept"].([]interface{}) 
	if ok {
		for _, v := range acceptArr {
			acc, ok := v.(string)
			if ok {
				err := accept.Add(acc)
				if err != nil {
					return nil, err
				}
			}

		}
		if len(acceptArr) == 0 {

			accept = nil
		}
	} else {
		accept = nil
	}

	// parse denied clients address if field not empty 
	deny, err := client.New()
	if err != nil {
		return nil, err
	}
	denyArr, ok := config["deny"].([]interface{})
	if ok {
		for _, v := range denyArr {
			den, ok := v.(string)
			if ok {
				err := deny.Add(den)
				if err != nil {
					return nil, err
				}
			}
		}
		if len(denyArr) == 0 {
			deny = nil
		}
	} else {
		deny = nil
	}

	// parse servers
	srvs := make([]*Server, 0)
	serversArr, ok := config["servers"].([]interface{})
	if ok {
		for i := 0; i < len(serversArr); i++ {
			serversStr, ok := serversArr[i].(map[string]interface{})
			if ok {
				srvss, err := ServersFromMap(serversStr, logDir)
				if err != nil {
					return nil, err
				}
				srvs = append(srvs, srvss...)
			}

		}
	}

	// parse servers. if field not empty use default 
	balancingStr, ok := config["balancing"].(string)
	if ok {
		balancingStr = ""
	}
	pool, err := NewPool(srvs, balancingStr)
	if err != nil {
		return nil, err
	}

	// read deadlines. if empty - no deadline
	var (
		dl, rdl, wdl, mconntime time.Duration
		maxconn                 int64
		toport                  int
	)
	if config["deadline"] != nil {
		dlS, ok := config["deadline"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler deadline %v", config["deadline"])
		}
		dl, err = time.ParseDuration(dlS)
		if err != nil {
			return nil, err
		}
	}
	if config["readdeadline"] != nil {
		rdlS, ok := config["readdeadline"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler readdeadline %v", config["readdeadline"])
		}
		rdl, err = time.ParseDuration(rdlS)
		if err != nil {
			return nil, err
		}
	}
	if config["writedeadline"] != nil {
		wdlS, ok := config["writedeadline"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler writedeadline %v", config["writedeadline"])
		}
		wdl, err = time.ParseDuration(wdlS)
		if err != nil {
			return nil, err
		}
	}
	// parse max time to connect server. if not exist infinity
	if config["maxconnectionstime"] != nil {
		mconntimeS, ok := config["maxconnectionstime"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler maxconnectionstime %v", config["maxconnectionstime"])
		}
		mconntime, err = time.ParseDuration(mconntimeS)
		if err != nil {
			return nil, err
		}
	}
	// parse max connections number. if not exist infinity
	if config["maxconnections"] != nil {
		maxconn, ok = config["maxconnections"].(int64)
		if !ok {
			return nil, fmt.Errorf("invalid handler maxconnections %v", config["maxconnections"])
		}
	}
	// parse destination port on servers
	if config["toport"] != nil {
		toport, ok = config["toport"].(int)
		if !ok {
			return nil, fmt.Errorf("invalid handler destiantion port %v", config["toport"])
		}
	}

	// parse ip filters (clients can be filtered by source address and they requests sends to different servers)
	filters := make([]*IPFilter, 0)
	filtersStr, ok := config["ipfilters"].([]map[string]interface{})
	if ok {
		for i := 0; i < len(filtersStr); i++ {
			srvs := make([]*Server, 0)
			serversStr, ok := filtersStr[i]["servers"].([]map[string]interface{})
			if ok {
				for i := 0; i < len(serversStr); i++ {
					srvss, err := ServersFromMap(serversStr[i], logDir)
					if err != nil {
						return nil, err
					}
					srvs = append(srvs, srvss...)
				}
			}
			balancingStr, ok := filtersStr[i]["balancing"].(string)
			if ok {
				balancingStr = ""
			}
			pool, err := NewPool(srvs, balancingStr)
			if err != nil {
				return nil, err
			}

			var source *client.Sources
			sourceStr, ok := filtersStr[i]["source"].([]string)
			if ok {
				source, err = client.New(sourceStr...)
				if err != nil {
					return nil, err
				}
			}
			filter := NewFilter(pool, source)
			filters = append(filters, filter)
		}

	}

	// create logger
	logFile := fmt.Sprintf("%s/%s_%s.log", logDir, protocol, port)
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, " ", log.LstdFlags)
	logger.SetFlags(log.LstdFlags)

	return &Handler{
		Protocol: protocol,
		Port:     port,
		logger:   logger,

		Toport:         toport,
		Accept:         accept,
		Deny:           deny,
		Servers:        pool,
		IPFilter:       filters,
		DeadLine:       dl,
		WriteDeadLine:  wdl,
		ReadDeadLine:   rdl,
		MaxConnectTime: mconntime,
		MaxConnections: maxconn,
	}, err

}

//	Run handler job gorutine
//
func (s *Handler) Listen() {
	go s.listen()
}

//	Get requests conn and delegate they to handle function
//
func (s *Handler) listen() error {
	listener, err := net.Listen(s.Protocol, fmt.Sprintf("0.0.0.0:%s", s.Port))
	if err != nil {
		return err
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.handle(conn)
	}

}

//
//
func (s *Handler) handle(client net.Conn) {
	// if max connections reached client will be wait or request will be rejected (reject default)
	atomic.AddUint64(&s.connectionsNumber, 1)
	if s.MaxConnections != 0 && atomic.LoadInt64(&s.currentconnectionsNumber) >= s.MaxConnections {
		switch s.OverFlow {
		case "wait":
			for {
				if atomic.LoadInt64(&s.currentconnectionsNumber) <= s.MaxConnections {
					break
				} else {
					time.Sleep(2 * time.Second)
				}
			}

		case "reject", "":
			client.Close()
			atomic.AddUint64(&s.rejected, 1)
			return
		}
	}

	//	Check is accepted client address
	//
	if s.Accept != nil && !s.Accept.Contains(client.RemoteAddr().String()) {
		client.Close()
		atomic.AddUint64(&s.rejected, 1)
		return
	} else if s.Deny.Contains(client.RemoteAddr().String()) {
		client.Close()
		atomic.AddUint64(&s.rejected, 1)
		return
	}
	atomic.AddInt64(&s.currentconnectionsNumber, 1)

	//	Check and set deadlines
	//
	start := time.Now()
	if s.DeadLine != 0 {
		client.SetDeadline(start.Add(s.DeadLine))
		fmt.Println("as")
	}
	if s.ReadDeadLine != 0 {
		client.SetReadDeadline(start.Add(s.ReadDeadLine))
	}
	if s.WriteDeadLine != 0 {
		client.SetWriteDeadline(start.Add(s.WriteDeadLine))
	}

	//	Compare client ip and servers
	srvpool := s.Servers
	for i := 0; i < len(s.IPFilter); i++ {
		pool := s.IPFilter[i].Contains(client.RemoteAddr().String())
		if pool != nil {
			srvpool = pool
			break
		}
	}

	//	Find available server and connect to they
	var err error
	var srv *Server
	var server net.Conn
	for i := 0; i < len(srvpool.Servers) && server == nil; i++ {
		srv, err = srvpool.FindServer(client.RemoteAddr().String())
		if err != nil {
			client.Close()
			fmt.Println(err)
			return
		}
		server, err = srv.Connect(s.Protocol, strconv.Itoa(s.Toport))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	srv.Exchange(client, server)

	atomic.AddInt64(&s.currentconnectionsNumber, -1)
}
