package ippool

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"
)

type BalancingMethod interface {
	FindServer(ip string, p []Server) (*Server, error)
	Rebalance([]Server)
}

func NewBalancingMethod(name string) (BalancingMethod, error) {
	switch name {
	case "roundRobin":
		return &RoundRobin{counter: 0, weightCounter: 1}, nil
	case "none":
		return &None{}, nil
	case "random":
		return &Random{weightMap: make(map[int]int)}, nil
	case "hashIP":
		return &HashIP{weightMap: make(map[int]int)}, nil
	case "leastConnections":
		return &LeastConnections{}, nil
	case "auto":
		return &Auto{}, nil
	default:
		return nil, nil
	}
}

type RoundRobin struct {
	counter       int
	weightCounter int
}

func (m *RoundRobin) FindServer(sIP string, p []Server) (*Server, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no servers avaible in pool")
	}
	var srv *Server
	if m.weightCounter == 1 {
		m.counter++
		srv = &p[m.counter%len(p)]
		m.weightCounter = srv.Weight
	} else {
		srv = &p[m.counter%len(p)]
		m.weightCounter--
	}
	return srv, nil
}

func (m *RoundRobin) Rebalance(p []Server) {
	if len(p) == 0 {
		return
	}
	if m.weightCounter > p[m.counter%len(p)].Weight {
		m.weightCounter = p[m.counter%len(p)].Weight
	}
}

type Random struct {
	weightMap map[int]int
}

func (m *Random) FindServer(sIP string, p []Server) (*Server, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no servers avaible in pool")
	}
	rand.Seed(time.Now().UnixNano())
	n := int(rand.Int63()) % len(m.weightMap)
	srv := &p[m.weightMap[n]]
	return srv, nil
}

func (m *Random) Rebalance(p []Server) {
	if len(p) == 0 {
		return
	}
	counter := 0
	for i := 0; i < len(p); i++ {
		for ii := p[i].Weight; ii > 0; ii-- {
			m.weightMap[counter] = i
		}
	}
}

type None struct {
}

func (m *None) FindServer(sIP string, p []Server) (*Server, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no servers avaible in pool")
	}
	srv := &p[0]
	return srv, nil
}

func (m *None) Rebalance(p []Server) {
	if len(p) == 0 {
		return
	}
	weight := 0
	highPriority := 0
	for i := 0; i < len(p); i++ {
		if weight < p[i].Weight {
			highPriority = i
		}
	}
	p[0], p[highPriority] = p[highPriority], p[0]
}

type LeastConnections struct {
}

func (m *LeastConnections) FindServer(sIP string, p []Server) (*Server, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no servers avaible in pool")
	}
	srv := &p[0]
	for i := 1; i < len(p); i++ {
		if p[i].CurrentConnectionsNumber < srv.CurrentConnectionsNumber {
			srv = &p[i]
		}
	}
	return srv, nil
}

func (m *LeastConnections) Rebalance(p []Server) {

}

type Auto struct {
}

func (m *Auto) FindServer(sIP string, p []Server) (*Server, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no servers avaible in pool")
	}
	srv := &p[0]
	prior := p[0].Priority * float64(p[0].CurrentConnectionsNumber)
	for i := 1; i < len(p); i++ {
		pr := p[i].Priority * float64(p[i].CurrentConnectionsNumber)
		if pr > prior {
			srv = &p[i]
		}
	}
	return srv, nil
}

func (m *Auto) Rebalance(p []Server) {
	if len(p) == 0 {
		return
	}
	for i := 1; i < len(p); i++ {
		p[i].Priority = float64(p[i].Weight / (int(p[i].stats.AvgConnectTime) * int(p[i].stats.AvgDataExchangeTime/100)))
	}
}

type HashIP struct {
	weightMap map[int]int
}

func (m *HashIP) FindServer(sIP string, p []Server) (*Server, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no servers avaible in pool")
	}
	h := fnv.New32a()
	h.Write([]byte(sIP))
	srv := &p[m.weightMap[int(h.Sum32())%len(m.weightMap)]]
	return srv, nil
}

func (m *HashIP) Rebalance(p []Server) {
	if len(p) == 0 {
		return
	}
	counter := 0
	for i := 0; i < len(p); i++ {
		for ii := p[i].Weight; ii > 0; ii-- {
			m.weightMap[counter] = i
		}
	}
}
