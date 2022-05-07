package porthndlr

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/averageNetAdmin/andproxy/cmd/ipdb"
	"github.com/averageNetAdmin/andproxy/cmd/ippool"
)

//
//	the struct that keep info about "microprogramm" that listen the port
//
type Handler struct {
	port     string
	protocol string
	config   *Config
	logger   *log.Logger
}

type Config struct {
	accept           ippool.Pool
	deny             ippool.Pool
	servers          ippool.ServerPool
	filter           ippool.Filter
	toport           string
	ServersWeight    int
	ServersMaxFails  int
	ServersBreakTime int
}

func NewConfig(hconf map[string]interface{}, db *ipdb.IPDB, port, handlerLogDir string) (*Config, error) {
	var (
		accept  ippool.Pool
		deny    ippool.Pool
		servers ippool.ServerPool
		filter  ippool.Filter
		toport  string
	)

	if valf, ok := hconf["accept"]; ok && valf != nil {
		if val, ok := valf.(string); ok && strings.HasPrefix(val, "$") {
			accept, ok = db.GetPoolCopy(val[1:])
			if !ok {
				return nil, fmt.Errorf("error: invalid accept pool name %v is not exist in port handler %v", val, port)
			}
		} else if val, ok := valf.([]string); ok {
			p := ippool.Pool{}
			err := p.AddArr(val)
			if err != nil {
				return nil, err
			}
			accept = p
		} else {
			return nil, fmt.Errorf("error: invalid accept pool %v in port handler %v", valf, port)
		}
	}

	if valf, ok := hconf["deny"]; ok && valf != nil {

		if val, ok := valf.(string); ok && strings.HasPrefix(val, "$") {
			deny, ok = db.GetPoolCopy(val[1:])
			if !ok {
				return nil, fmt.Errorf("error: invalid deny pool name %v is not exist in port handler %v", val, port)
			}
		} else if val, ok := valf.([]string); ok {
			p := ippool.Pool{}
			err := p.AddArr(val)
			if err != nil {
				return nil, err
			}
			deny = p
		} else {
			return nil, fmt.Errorf("error: invalid deny pool %v in port handler %v", valf, port)
		}
	}

	if valf, ok := hconf["servers"]; ok && valf != nil {
		if val, ok := valf.(string); ok && strings.HasPrefix(val, "$") {
			servers, ok = db.GetServerPoolCopy(val[1:])
			if !ok {
				return nil, fmt.Errorf("error: invalid server pool name %v is not existin port handler %v", val, port)
			}
		} else if val, ok := valf.(map[string]interface{}); ok {
			p := ippool.ServerPool{}
			err := p.AddFromMap(val)
			if err != nil {
				return nil, err
			}
			servers = p
		} else {
			return nil, fmt.Errorf("error: invalid server pool %v in port handler %v", valf, port)
		}
	}
	err := servers.SetLogFile(handlerLogDir + "/servers")

	if err != nil {
		return nil, err
	}
	if valf, ok := hconf["filters"]; ok && valf != nil {
		if val, ok := valf.(string); ok && strings.HasPrefix(val, "$") {
			filter, ok = db.GetFilterCopy(val[1:])
			if !ok {
				return nil, fmt.Errorf("error: filter %v in port handler %s is exist", filter, port)
			}
		} else {
			return nil,
				fmt.Errorf("error: invalid filter name %v in port handler, filter must be in 'filters' in port handler %s", valf, port)
		}
	}

	if err != nil {
		return nil, err
	}
	if valf, ok := hconf["toport"]; ok && valf != nil {
		val, ok := valf.(int)
		if !ok {
			return nil, fmt.Errorf("error: invalid port number %v in port hangler %s", valf, port)
		}
		if val < 0 || val > 65535 {
			return nil, fmt.Errorf("error: invalid port number %d", val)
		}
		toport = strconv.Itoa(val)
	} else {
		toport = port
	}

	if valf, ok := hconf["balancingMethod"]; ok && valf != nil {
		val, ok := valf.(string)
		if !ok {
			return nil, fmt.Errorf("error: invalid balancing method %v in port handler %s", valf, port)
		}

		err := servers.SetBalancingMethod(val)
		if err != nil {
			return nil, err
		}
		err = filter.SetBalancingMethod(val)
		if err != nil {
			return nil, err
		}
	}
	h := &Config{
		accept:  accept,
		deny:    deny,
		servers: servers,
		filter:  filter,
		toport:  toport,
	}
	return h, nil

}

//
//	the struct that keep info about "microprogramm" that listen the port
//
func NewHandler(protocol, port string, db *ipdb.IPDB, hconf map[string]interface{}, logDir string) (*Handler, error) {

	handlderLogDir := logDir + "/" + protocol + "_" + port
	err := os.MkdirAll(handlderLogDir, 0644)
	if err != nil {
		return nil, err
	}
	handlerLogFile := handlderLogDir + "/handler.log"
	file, err := os.OpenFile(handlerLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	config, err := NewConfig(hconf, db, port, handlderLogDir)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, "andproxy server log: ", log.LstdFlags)
	logger.SetFlags(log.LstdFlags)
	NewConfig(hconf, db, port, handlderLogDir)
	h := &Handler{
		port:     port,
		protocol: protocol,
		config:   config,
		logger:   logger,
	}
	return h, nil

}

func (h *Handler) Handle() {
	go h.handle()
}

//
//	start listenting the port
//
func (h *Handler) handle() error {
	server, err := net.Listen(h.protocol, "0.0.0.0:"+h.port)
	fmt.Println("aaa")
	if err != nil {
		return err
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			return err
		}
		clientAddress, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			return err
		}
		v, err := h.config.filter.WhatPool(clientAddress)
		if err != nil {
			return err
		}
		if !h.config.accept.Empty() {
			if yes, _ := h.config.accept.Contains(clientAddress); !yes {
				h.logger.Printf("acccess denied to %v because it not in accept list\n", clientAddress)
				continue
			}
		} else if !h.config.deny.Empty() {
			if yes, _ := h.config.deny.Contains(clientAddress); yes {
				h.logger.Printf("acccess denied to %v because it in deny list\n", clientAddress)
				continue
			}
		}
		fmt.Println(conn)
		h.logger.Printf("successfull connection from %v \n", clientAddress)
		if v != nil {
			go h.exchangeData(clientAddress, v, conn)
		} else {
			go h.exchangeData(clientAddress, &h.config.servers, conn)
		}
	}
}

func (h *Handler) exchangeData(clientAddr string, srvPool *ippool.ServerPool, clientConn net.Conn) {
	fmt.Println(clientConn)
	srv, err := srvPool.FindServer(clientAddr)
	if err != nil {
		h.logger.Println(err)
		return
	}
	srvConn, err := srv.Connect(h.protocol, h.port)
	if err != nil {
		h.logger.Println(err)
		srvPool.UpdateBroken()
		for i := 0; i < len(srvPool.Servers)*srv.MaxFails; i++ {
			srv, err := srvPool.FindServer(clientAddr)
			if err != nil {
				return
			}
			srvConn, err = srv.Connect(h.protocol, h.port)
			if err == nil {
				break
			}

		}
	}
	defer srv.Disconnect(srvConn)
	defer clientConn.Close()
	srv.ExchangeData(clientConn, srvConn)

}

func (h *Handler) UpdateConfig(c *Config) {
	h.config = c
	fmt.Println(h.config.servers.Servers)
}
