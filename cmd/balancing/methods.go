package balancing

import (
	"github.com/averageNetAdmin/andproxy/cmd/ippool"
)

type Method interface {
	FindServer(ip string, p *ippool.ServerPool) (*ippool.Server, error)
	FindAnotherServer(ip string, p *ippool.ServerPool) (*ippool.Server, error)
}

func NewMethod(name string) (Method, error) {
	switch name {
	case "roundRobin":
		return &RoundRobin{counter: 0}, nil
	default:
		return nil, nil
	}
}

type RoundRobin struct {
	counter int
}

func (m *RoundRobin) FindServer(sIP string, p *ippool.ServerPool) (*ippool.Server, error) {
	srv := &p.Servers[m.counter%len(p.Servers)]
	m.counter++
	return srv, nil
}

func (m *RoundRobin) FindAnotherServer(sIP string, p *ippool.ServerPool) (*ippool.Server, error) {
	srv := &p.Servers[m.counter%len(p.Servers)]
	m.counter++
	return srv, nil
}
