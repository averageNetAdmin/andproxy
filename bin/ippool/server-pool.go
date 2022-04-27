package ippool

import (
	"net/netip"
	"strings"
)

//
//	Find - return the server that should handle request
//
type BalancingMethod interface {
	Find(ip string, p *ServerPool) (string, error)
}

//
//	the struct that keep servers addresses
//	contain only IP addresses
//
type ServerPool struct {
	addr []netip.Addr
}

//
//	check that server should handle request
//
func (p *ServerPool) WhatServer(ip string, bm BalancingMethod) (string, error) {
	add, err := bm.Find(ip, p)
	if err != nil {
		return "", err
	}
	return add, nil
}

//
//	display all servers IP
//
func (p *ServerPool) String() string {
	result := ""
	for _, addr := range p.addr {
		result += addr.String() + " "
	}
	return result
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
	for _, addr := range p.addr {
		if addr.Compare(ip) == 0 {
			return true, nil
		}
	}
	return false, nil
}

//
//	add IP address in servers pool
//
func (p *ServerPool) Add(a string) error {
	isRange := strings.Contains(a, "-")
	if isRange {
		rng, err := genIPRange(a)
		if err != nil {
			return err
		}
		p.addr = append(p.addr, rng...)
	} else {
		addr, err := netip.ParseAddr(a)
		if err != nil {
			return err
		}
		p.addr = append(p.addr, addr)
	}
	return nil
}

//
//	add IP address in servers pool from array
//
func (p *ServerPool) AddArr(arr []string) error {
	for _, elem := range arr {
		err := p.Add(elem)
		if err != nil {
			return err
		}
	}
	return nil
}

//
//	create and return new ServerPool
//
func NewServerPool(ip ...string) (*ServerPool, error) {
	p := new(ServerPool)
	err := p.AddArr(ip)
	if err != nil {
		return nil, err
	}
	return p, nil
}
