package ippool

import (
	"fmt"
	"net"
	"net/netip"
	"strings"
)

//
//	the struct that keep servers
//
type ServerPool struct {
	Name    string
	Servers []Server
	Broken  []Server
	BM      BalancingMethod
}

//
//	if server is broken - it move to array of broken server.
//	if server returns to work - it move to array of working servers
//
func (s *ServerPool) UpdateBroken() {
	for i := 0; i < len(s.Servers); i++ {
		if s.Servers[i].Broken {
			s.Broken = append(s.Broken, s.Servers[i])
			s.Servers = append(s.Servers[:i], s.Servers[i+1:]...)
		}
	}
	for i := 0; i < len(s.Broken); i++ {
		if !s.Broken[i].Broken {
			s.Servers = append(s.Servers, s.Broken[i])
			s.Broken = append(s.Broken[:i], s.Broken[i+1:]...)
		}
	}
	s.BM.Rebalance(s.Servers)
}

//
//	set balancing method
//
func (s *ServerPool) SetBalancingMethod(name string) error {
	bm, err := NewBalancingMethod(name)
	if err != nil {
		return err
	}
	s.BM = bm
	return nil
}

//
//	find server in pool. for finding uses currnet balancing method
//
func (s *ServerPool) FindServer(ip string) (*Server, error) {
	srv, err := s.BM.FindServer(ip, s.Servers)
	if err != nil {
		return nil, err
	}
	return srv, nil
}

//
//	rebalance servers pool
//
func (s *ServerPool) Rebalance() {
	s.BM.Rebalance(s.Servers)
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
//	add server in servers pool
//
func (p *ServerPool) Add(serverNames string, config map[string]interface{}) error {

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
	isRange := strings.Contains(serverNames, "-")
	if isRange {
		rng, err := CreateIPRange(serverNames)
		if err == nil {
			for _, addr := range rng {
				srv := NewServer(addr.String(), weight, maxFails, breakTime)
				p.Servers = append(p.Servers, *srv)
			}
		} else {
			rng, err := CreateRange(serverNames)
			if err != nil {
				return err
			}
			for _, srvNames := range rng {
				_, err := net.LookupHost(srvNames)
				if err != nil {
					return err
				}
				srv := NewServer(srvNames, weight, maxFails, breakTime)
				p.Servers = append(p.Servers, *srv)
			}
		}

	} else {
		addr, err := netip.ParseAddr(serverNames)
		if err == nil {
			srv := NewServer(addr.String(), weight, maxFails, breakTime)
			p.Servers = append(p.Servers, *srv)
		} else {
			_, err := net.LookupHost(serverNames)
			if err != nil {
				return err
			}

			srv := NewServer(serverNames, weight, maxFails, breakTime)
			p.Servers = append(p.Servers, *srv)
		}

	}
	return nil
}

//
//	add servers in servers pool from map
//
func (p *ServerPool) AddFromMap(m map[string]interface{}) error {
	for addresses, config := range m {
		if addresses == "balancingMethod" {
			err := p.SetBalancingMethod(addresses)
			if err != nil {
				return err
			}
			continue
		}
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
	p.SetBalancingMethod("roundRobin")
	err := p.AddFromMap(m)
	if err != nil {
		return nil, err
	}
	return p, nil
}

//
//	set new config to all servers in pool
//
func (p *ServerPool) SetConfig(weight, maxFails, breakTime string) error {
	for i := 0; i < len(p.Servers); i++ {
		err := p.Servers[i].SetConfig(weight, maxFails, breakTime)
		if err != nil {
			return err
		}
	}
	return nil
}
