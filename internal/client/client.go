package client

import (
	"net"
	"net/netip"
	"strings"

	"github.com/averageNetAdmin/andproxy/internal/ranges"
)

// Contain pool of addresses and pool of networks
//
type Sources struct {
	Addrs []netip.Addr
	Nets  []netip.Prefix
}

//	Check is ip address in struct
//	Return true if struct contains searchIP ip address else return false
//	If searchIP is not valid ip address return false
//
func (s *Sources) Contains(searchIP string) bool {
	host, _, _ := net.SplitHostPort(searchIP)
	pHost, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	for _, a := range s.Addrs {
		if a.Compare(pHost) == 0 {
			return true
		}
	}
	for _, n := range s.Nets {
		if n.Contains(pHost) {
			return true
		}
	}
	return false
}

//	Add net, address, net range or address range
//
func (s *Sources) Add(addrs string) error {
	rng, err := ranges.Create(addrs)
	if err != nil {
		return err
	}
	for _, el := range rng {
		if strings.Contains(el, "/") {
			net, err := netip.ParsePrefix(el)
			if err != nil {
				return err
			}
			s.Nets = append(s.Nets, net)
		} else {
			addr, err := netip.ParseAddr(el)
			if err != nil {
				return err
			}
			s.Addrs = append(s.Addrs, addr)
		}
	}

	return nil
}

//	Use to add addresses or nets from array
//
func (ar *Sources) AddFromArr(arr []string) error {
	for _, elem := range arr {
		err := ar.Add(elem)
		if err != nil {
			return err
		}
	}
	return nil
}

// Return new Sources from ip range
//
func New(ip ...string) (*Sources, error) {
	p := new(Sources)
	return p, p.AddFromArr(ip)
}
