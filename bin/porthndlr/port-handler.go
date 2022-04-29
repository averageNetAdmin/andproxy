package porthndlr

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/averageNetAdmin/andproxy/bin/ipdb"
	"github.com/averageNetAdmin/andproxy/bin/ippool"
	"github.com/averageNetAdmin/andproxy/bin/srcfltr"
)

//
//	the struct that keep info about "microprogramm" that listen the port
//
type PortHandler struct {
	port            string
	protocol        string
	accept          *ippool.Pool
	deny            *ippool.Pool
	servers         *ippool.ServerPool
	filter          *srcfltr.Filter
	toport          string
	balancingMethod string
}

//
//	the struct that keep info about "microprogramm" that listen the port
//
func NewPortHandler(protocol, port string, db *ipdb.IPDB, hconf map[string]interface{}) (*PortHandler, error) {

	var (
		accept          *ippool.Pool
		deny            *ippool.Pool
		servers         *ippool.ServerPool
		filter          *srcfltr.Filter
		toport          string
		balancingMethod string
	)

	if valf, ok := hconf["accept"]; ok && valf != nil {
		val := valf.(string)
		fmt.Println(val)
		if strings.HasPrefix(val, "$") {
			accept = db.GetPool(val[1:])
			if accept == nil {
				return nil, errors.New("Pool " + val + " not found")
			}
		} else {
			p := new(ippool.Pool)
			err := p.Add(val)
			if err != nil {
				return nil, err
			}
			accept = p
		}
	} else {
		accept = nil
	}

	if valf, ok := hconf["deny"]; ok && valf != nil {
		val := valf.(string)
		if strings.HasPrefix(val, "$") {
			deny = db.GetPool(val[1:])
			if deny == nil {
				return nil, errors.New("Pool " + val + " not found")
			}
		} else {
			p := new(ippool.Pool)
			err := p.Add(val)
			if err != nil {
				return nil, err
			}
			deny = p
		}
	} else {
		deny = nil
	}

	if valf, ok := hconf["servers"]; ok && valf != nil {
		val := valf.(string)
		if strings.HasPrefix(val, "$") {
			servers = db.GetServerPool(val[1:])
			if servers == nil {
				return nil, errors.New("Pool " + val + " not found")
			}
		} else {
			p := new(ippool.ServerPool)
			err := p.Add(val)
			if err != nil {
				return nil, err
			}
			servers = p
		}
	} else {
		servers = nil
	}

	if valf, ok := hconf["filter"]; ok && valf != nil {
		val := valf.(string)
		if strings.HasPrefix(val, "$") {
			filter = db.GetFilter(val[1:])
			if filter == nil {
				return nil, errors.New("Pool " + val + " not found")
			}
		} else {
			vals := strings.Split(val, ":")
			p := new(srcfltr.Filter)
			ipp := new(ippool.Pool)
			err := ipp.Add(vals[0])
			if err != nil {
				return nil, err
			}
			srvp := new(ippool.ServerPool)
			err = srvp.Add(vals[1])
			if err != nil {
				return nil, err
			}
			p.Add(ipp, srvp)
			filter = p
		}
	} else {
		filter = nil
	}

	if valf, ok := hconf["toport"]; ok && valf != nil {
		val, ok := valf.(int)
		if !ok {
			return nil, errors.New("Invalid port number " + valf.(string))
		}
		if val < 0 || val > 65535 {
			return nil, errors.New("Invalid port number " + strconv.Itoa(val))
		}
		toport = strconv.Itoa(val)
	} else {
		toport = port
	}

	balancingMethod = "none"
	//
	// TODO: add balancing methods
	//

	h := &PortHandler{
		port:            port,
		protocol:        protocol,
		accept:          accept,
		deny:            deny,
		servers:         servers,
		filter:          filter,
		toport:          toport,
		balancingMethod: balancingMethod,
	}
	return h, nil

}

//
//	start listenting the port
//
func (h *PortHandler) Handle() error {
	fmt.Println(h.protocol)
	server, err := net.Listen(h.protocol, "0.0.0.0:"+h.port)
	if err != nil {
		return err
	}
	for {
		conn, err := server.Accept()
		fmt.Println("conn")
		if err != nil {
			return err
		}

		sourceAddress, _, err := net.SplitHostPort(conn.RemoteAddr().String())
		fmt.Println(sourceAddress)
		if err != nil {
			return err
		}
		data := make([]byte, 1500)
		n, err := conn.Read(data)
		if err != nil {
			return err
		}
		data = data[:n]
		var localConn net.Conn
		v, err := h.filter.WhatPool(sourceAddress)
		if err != nil {
			return err
		}

		if v != nil {
			fmt.Println(h.toport)
			ipa := *h.servers.Servers[0].String()
			localConn, err = net.Dial(h.protocol, ipa+":"+h.toport)
			if err != nil {
				return err
			}
		} else {
			ipa := h.servers.Servers[0].String()
			localConn, err = net.Dial(h.protocol, ipa+":"+h.toport)
			if err != nil {
				return err
			}
		}
		localConn.Write(data)
		datas := make([]byte, 1500)
		nn, err := localConn.Read(datas)
		if err != nil {
			return err
		}
		conn.Write(datas[:nn])
		localConn.Close()
		conn.Close()
	}
}
