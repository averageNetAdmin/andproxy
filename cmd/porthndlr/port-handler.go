package porthndlr

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/averageNetAdmin/andproxy/cmd/ipdb"
	"github.com/averageNetAdmin/andproxy/cmd/ippool"
)

//
//	the struct that keep info about "microprogramm" that listen the port
//
type Handler struct {
	Port           string
	Protocol       string
	Config         *Config
	ConnNumber     uint64
	DeniedConnNumb uint64
	Run            bool
	logger         *log.Logger
}

type Config struct {
	Accept           ippool.Pool
	Deny             ippool.Pool
	Servers          ippool.ServerPool
	Filter           ippool.Filter
	Toport           string
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
			p, err := ippool.NewServerPool(val)
			if err != nil {
				return nil, err
			}
			servers = *p
		} else {
			return nil, fmt.Errorf("error: invalid server pool %v in port handler %v", valf, port)
		}
	}
	err := servers.SetLogFile(fmt.Sprintf("%s/servers", handlerLogDir))

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

	if valf, ok := hconf["balancingmethod"]; ok && valf != nil {
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
	servers.Rebalance()
	filter.Rebalance()
	h := &Config{
		Accept:  accept,
		Deny:    deny,
		Servers: servers,
		Filter:  filter,
		Toport:  toport,
	}
	return h, nil

}

//
//	the struct that keep info about "microprogramm" that listen the port
//
func NewHandler(protocol, port string, db *ipdb.IPDB, hconf map[string]interface{}, logDir string) (*Handler, error) {

	handlderLogDir := fmt.Sprintf("%s/%s_%s", logDir, protocol, port)
	err := os.MkdirAll(handlderLogDir, 0644)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(fmt.Sprintf("%s/handler.log", handlderLogDir), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	config, err := NewConfig(hconf, db, port, handlderLogDir)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, fmt.Sprintf("handler %s_%s", protocol, port), log.LstdFlags)
	logger.SetFlags(log.LstdFlags)
	NewConfig(hconf, db, port, handlderLogDir)
	h := &Handler{
		Port:     port,
		Protocol: protocol,
		Config:   config,
		logger:   logger,
	}
	return h, nil

}

func (h *Handler) Handle() {
	h.Run = true
	go h.handle()
}

//
//	start listenting the port
//
func (h *Handler) handle() error {
	server, err := net.Listen(h.Protocol, net.JoinHostPort("0.0.0.0", h.Port))
	if err != nil {
		return err
	}
	for {
		if !h.Run {
			return nil
		}
		conn, err := server.Accept()
		if err != nil {
			return err
		}
		h.ConnNumber++
		clientAddress, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		if err != nil {
			return err
		}
		v, err := h.Config.Filter.WhatPool(clientAddress)
		if err != nil {
			return err
		}
		if !h.Config.Accept.Empty() {
			if yes, _ := h.Config.Accept.Contains(clientAddress); !yes {
				h.logger.Printf("acccess denied to %v because it not in accept list\n", clientAddress)
				h.DeniedConnNumb++
				continue
			}
		} else if !h.Config.Deny.Empty() {
			if yes, _ := h.Config.Deny.Contains(clientAddress); yes {
				h.logger.Printf("acccess denied to %v because it in deny list\n", clientAddress)
				h.DeniedConnNumb++
				continue
			}
		}
		h.logger.Printf("successfull connection from %v \n", clientAddress)
		if v != nil {
			go h.exchangeData(clientAddress, v, conn)
		} else {
			go h.exchangeData(clientAddress, &h.Config.Servers, conn)
		}
	}
}

func (h *Handler) exchangeData(clientAddr string, srvPool *ippool.ServerPool, clientConn net.Conn) {

	srv, err := srvPool.FindServer(clientAddr)
	if err != nil {
		h.logger.Println(err)
		return
	}
	srvConn, err := srv.Connect(h.Protocol, h.Config.Toport)
	if err != nil {
		h.logger.Println(err)
		srvPool.UpdateBroken()
		for i := 0; i < len(srvPool.Servers)*srv.MaxFails; i++ {
			srv, err := srvPool.FindServer(clientAddr)
			if err != nil {
				return
			}
			srvConn, err = srv.Connect(h.Protocol, h.Config.Toport)
			if err == nil {
				break
			}
		}
		if srvConn == nil {
			return
		}
	}
	srv.ExchangeData(clientConn, srvConn)
	srv.Disconnect(srvConn)
	clientConn.Close()
}

func (h *Handler) UpdateConfig(c *Config) {
	h.Config = c
}

func (h *Handler) Stop() {
	h.Run = false
}

func (h *Handler) SaveState() error {
	dirName := fmt.Sprintf("/var/lib/andproxy/states/%s_%s", h.Protocol, h.Port)
	t := time.Now()
	fileName := fmt.Sprintf("%s/%s", dirName, t.Format("2006-01-02T15:04:05"))
	err := os.MkdirAll(dirName, 0700)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0700)
	if err != nil {
		return err
	}
	info, err := json.Marshal(h)
	if err != nil {
		return err
	}
	_, err = file.Write(info)
	if err != nil {
		return err
	}
	return nil
}
