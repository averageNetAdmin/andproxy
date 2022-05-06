package porthndlr

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/averageNetAdmin/andproxy/cmd/balancing"

	"github.com/averageNetAdmin/andproxy/cmd/ipdb"
	"github.com/averageNetAdmin/andproxy/cmd/ippool"
)

//
//	the struct that keep info about "microprogramm" that listen the port
//
type PortHandler struct {
	port             string
	protocol         string
	accept           ippool.Pool
	deny             ippool.Pool
	servers          ippool.ServerPool
	filter           ippool.Filter
	toport           string
	balancingMethod  balancing.Method
	ServersWeight    int
	ServersMaxFails  int
	ServersBreakTime int
	logger           *log.Logger
}

//
//	the struct that keep info about "microprogramm" that listen the port
//
func NewPortHandler(protocol, port string, db *ipdb.IPDB, hconf map[string]interface{}, logDir string) (*PortHandler, error) {

	var (
		accept          ippool.Pool
		deny            ippool.Pool
		servers         ippool.ServerPool
		filter          ippool.Filter
		toport          string
		balancingMethod balancing.Method
	)
	handlderLogDir := logDir + "/" + protocol + "_" + port
	err := os.MkdirAll(handlderLogDir, 0644)
	if err != nil {
		return nil, err
	}
	if valf, ok := hconf["accept"]; ok && valf != nil {
		if val, ok := valf.(string); ok && strings.HasPrefix(val, "$") {
			accept, ok = db.GetPoolCopy(val[1:])
			if !ok {
				return nil, fmt.Errorf("error: invalid accept pool name %v is not exist in port handler %v %v", val, protocol, port)
			}
		} else if val, ok := valf.([]string); ok {
			p := ippool.Pool{}
			err := p.AddArr(val)
			if err != nil {
				return nil, err
			}
			accept = p
		} else {
			return nil, fmt.Errorf("error: invalid accept pool %v in port handler %v %v", valf, protocol, port)
		}
	}

	if valf, ok := hconf["deny"]; ok && valf != nil {

		if val, ok := valf.(string); ok && strings.HasPrefix(val, "$") {
			deny, ok = db.GetPoolCopy(val[1:])
			if !ok {
				return nil, fmt.Errorf("error: invalid deny pool name %v is not exist in port handler %v %v", val, protocol, port)
			}
		} else if val, ok := valf.([]string); ok {
			p := ippool.Pool{}
			err := p.AddArr(val)
			if err != nil {
				return nil, err
			}
			deny = p
		} else {
			return nil, fmt.Errorf("error: invalid deny pool %v in port handler %v %v", valf, protocol, port)
		}
	}

	if valf, ok := hconf["servers"]; ok && valf != nil {
		if val, ok := valf.(string); ok && strings.HasPrefix(val, "$") {
			servers, ok = db.GetServerPoolCopy(val[1:])
			if !ok {
				return nil, fmt.Errorf("error: invalid server pool name %v is not existin port handler %v %v", val, protocol, port)
			}
		} else if val, ok := valf.(map[string]interface{}); ok {
			p := ippool.ServerPool{}
			err := p.AddFromMap(val)
			if err != nil {
				return nil, err
			}
			servers = p
		} else {
			return nil, fmt.Errorf("error: invalid server pool %v in port handler %v %v", valf, protocol, port)
		}
	}
	err = servers.SetLogFile(handlderLogDir + "/servers")

	if err != nil {
		return nil, err
	}
	if valf, ok := hconf["filters"]; ok && valf != nil {
		if val, ok := valf.(string); ok && strings.HasPrefix(val, "$") {
			filter, ok = db.GetFilterCopy(val[1:])
			if !ok {
				return nil, fmt.Errorf("error: filter %v in port handler %s %s is exist", filter, protocol, port)
			}
		} else {
			return nil,
				fmt.Errorf("error: invalid filter name %v in port handler, filter must be in 'filters' in port handler %s %s", valf, protocol, port)
		}
	}

	if err != nil {
		return nil, err
	}
	if valf, ok := hconf["toport"]; ok && valf != nil {
		val, ok := valf.(int)
		if !ok {
			return nil, fmt.Errorf("error: invalid port number %v in port hangler %s %s", valf, protocol, port)
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
			return nil, fmt.Errorf("error: invalid balancing method %v in port handler %s %s", valf, protocol, port)
		}
		var err error
		balancingMethod, err = balancing.NewMethod(val)
		if err != nil {
			return nil, err
		}
	} else {
		balancingMethod, _ = balancing.NewMethod("roundRobin")
	}

	if err != nil {
		return nil, err
	}
	handlerLogFile := handlderLogDir + "/handler.log"
	file, err := os.OpenFile(handlerLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, "andproxy server log: ", log.LstdFlags)
	logger.SetFlags(log.LstdFlags)

	h := &PortHandler{
		port:            port,
		protocol:        protocol,
		accept:          accept,
		deny:            deny,
		servers:         servers,
		filter:          filter,
		toport:          toport,
		balancingMethod: balancingMethod,
		logger:          logger,
	}
	return h, nil

}

func (h *PortHandler) Handle() {
	go h.handle()
}

//
//	start listenting the port
//
func (h *PortHandler) handle() error {
	server, err := net.Listen(h.protocol, "0.0.0.0:"+h.port)
	if err != nil {
		return err
	}
	for {
		conn, err := server.Accept()
		if err != nil {
			return err
		}
		fmt.Println("aaa")
		clientAddress, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			return err
		}
		v, err := h.filter.WhatPool(clientAddress)
		if err != nil {
			return err
		}
		if !h.accept.Empty() {
			if yes, _ := h.accept.Contains(clientAddress); !yes {
				h.logger.Printf("acccess denied to %v because it not in accept list\n", clientAddress)
				continue
			}
		} else if !h.deny.Empty() {
			if yes, _ := h.deny.Contains(clientAddress); yes {
				h.logger.Printf("acccess denied to %v because it in deny list\n", clientAddress)
				continue
			}
		}
		fmt.Println(conn)
		h.logger.Printf("successfull connection from %v \n", clientAddress)
		if v != nil {
			go h.exchangeData(clientAddress, v, conn)
		} else {
			go h.exchangeData(clientAddress, &h.servers, conn)
		}
		fmt.Println("bbb")
	}
}

func (h *PortHandler) exchangeData(clientAddr string, srvPool *ippool.ServerPool, clientConn net.Conn) {
	fmt.Println(clientConn)
	srv, err := h.balancingMethod.FindServer(clientAddr, srvPool)
	if err != nil {
		h.logger.Println(err)
		return
	}
	srvConn, err := srv.Connect(h.protocol, h.port)
	if err != nil {
		h.logger.Println(err)
		for i := 0; i < len(srvPool.Servers); i++ {
			srv, err := h.balancingMethod.FindAnotherServer(clientAddr, srvPool)
			if err != nil {
				return
			}
			srvConn, err = srv.Connect(h.protocol, h.port)
			if err == nil {
				break
			}

		}
	}
	srv.ExchangeData(clientConn, srvConn)
	fmt.Println("here")
	srv.Disconnect(srvConn)
	clientConn.Close()
}
