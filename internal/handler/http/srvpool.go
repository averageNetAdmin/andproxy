package http

import (
	"github.com/averageNetAdmin/andproxy/internal/balancing"
)

type Pool struct {
	Servers   []*Server
	Broken    []*Server
	balancing balancing.Method
}

func NewPool(servers []*Server, balancingMethod string) (*Pool, error) {

	var bm balancing.Method
	broken := make([]*Server, 0)
	var err error
	if balancingMethod == "" {
		bm, err = balancing.NewMethod("roundrobin")
	} else {
		bm, err = balancing.NewMethod(balancingMethod)
	}
	if err != nil {
		return nil, err
	}
	srvs := make([]balancing.BalanceItem, 0)
	for i := 0; i < len(servers); i++ {
		srvs = append(srvs, servers[i])
	}
	bm.Rebalance(srvs)
	return &Pool{
		Servers:   servers,
		Broken:    broken,
		balancing: bm,
	}, nil
}

func (p *Pool) UpdateBroken() {
	for i := 0; i < len(p.Servers); i++ {
		if p.Servers[i].broken {
			p.Broken = append(p.Broken, p.Servers[i])
			p.Servers = append(p.Servers[:i], p.Servers[i+1:]...)
		}
	}
	for i := 0; i < len(p.Broken); i++ {
		if !p.Broken[i].broken {
			p.Servers = append(p.Servers, p.Broken[i])
			p.Broken = append(p.Broken[:i], p.Broken[i+1:]...)
		}
	}
	srvs := make([]balancing.BalanceItem, 0)
	for i := 0; i < len(p.Servers); i++ {
		srvs = append(srvs, p.Servers[i])
	}
	p.balancing.Rebalance(srvs)
}

func (s *Pool) FindServer(ip string) (*Server, error) {
	srvs := make([]balancing.BalanceItem, 0)
	for i := 0; i < len(s.Servers); i++ {
		srvs = append(srvs, s.Servers[i])
	}
	srv, err := s.balancing.FindServer(ip, srvs)
	if err != nil {
		return nil, err
	}
	srvv := srv.(*Server)
	return srvv, nil
}
