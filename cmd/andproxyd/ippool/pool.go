package ippool

import (
	"fmt"
	"net/netip"
	"strings"
)

//
//	the struct that keep clients addresses
//	caontain IP addresses "192.168.0.1"
// 	and IP networks "192.168.0.0/24"
//
type Pool struct {
	addr []netip.Addr
	nets []netip.Prefix
}

//
//	display all IP addresses and networks
//
func (ar *Pool) String() string {
	result := "["
	for _, addr := range ar.addr {
		result += addr.String() + " "
	}
	for _, s := range ar.nets {
		result += s.String() + " "
	}
	return result + "]"
}

//
//	check the IP addresse in pool or not
//	return true if pool contain IP or pool contain networm what contain ip
//
func (p *Pool) Contains(searchIP string) (bool, error) {
	sIP, err := netip.ParseAddr(searchIP)
	if err != nil {
		return false, err
	}
	for _, a := range p.addr {
		if a.Compare(sIP) == 0 {
			return true, nil
		}
	}
	for _, n := range p.nets {
		if n.Contains(sIP) {
			return true, nil
		}
	}
	return false, nil
}

//
//	add IP or IP range or network to the pool
//
func (p *Pool) Add(ip string) error {
	isRange := strings.Contains(ip, "-")
	isNet := strings.Contains(ip, "/")
	if isRange && isNet {
		return fmt.Errorf("invalid address %v\n Address cannot contain Net and Range at the same time", ip)
	} else if isRange {
		rng, err := CreateIPRange(ip)
		if err != nil {
			return err
		}
		p.addr = append(p.addr, rng...)
	} else if isNet {
		net, err := netip.ParsePrefix(ip)
		if err != nil {
			return err
		}
		p.nets = append(p.nets, net)
	} else {
		addr, err := netip.ParseAddr(ip)
		if err != nil {
			return err
		}
		p.addr = append(p.addr, addr)
	}
	return nil
}

//
//	add IP or IP range or network to the pool from range of strings
//
func (ar *Pool) AddArr(arr []string) error {
	for _, elem := range arr {
		err := ar.Add(elem)
		if err != nil {
			return err
		}
	}
	return nil
}

//
//	create and return new Pool
//
func NewPool(ip ...string) (*Pool, error) {
	p := new(Pool)
	err := p.AddArr(ip)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Pool) Empty() bool {
	if len(p.addr) == 0 && len(p.nets) == 0 {
		return true
	}
	return false
}
