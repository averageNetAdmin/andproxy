package srv

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"sync"
	"time"
)

// Interface to balancing method
type BalancingMethod interface {
	Rebalance([]int64) []int64
	FindServer(string, []int64) (int64, error)
}

// return new balancing method with checked name
func NewMethod(name string) (BalancingMethod, error) {
	switch name {
	case "roundrobin":
		return &RoundRobin{counter: 0, weightCounter: 1}, nil
	case "none":
		return &None{}, nil
	case "random":
		return &Random{weightMap: make(map[int]int)}, nil
	case "haship":
		return &HashIP{weightMap: make(map[int]int)}, nil
	// case "leastconnections":
	// 	return &LeastConnections{}, nil
	default:
		return nil, fmt.Errorf("%s balancing method not exist", name)
	}
}

// Filter requests by client ip address hash
// Client always go to same server
type HashIP struct {
	weightMap map[int]int
	mu        sync.RWMutex
}

func (m *HashIP) FindServer(sIP string, p []int64) (int64, error) {
	if len(p) == 0 {
		return 0, fmt.Errorf("no server.Servers avaible in pool")
	}
	h := fnv.New32a()
	h.Write([]byte(sIP))
	m.mu.RLock()
	srv := p[m.weightMap[int(h.Sum32())%len(m.weightMap)]]
	m.mu.RUnlock()
	return srv, nil
}

// If count of servers was changed, weight map must be changed
func (m *HashIP) Rebalance(p []int64) []int64 {
	if len(p) == 0 {
		return nil
	}
	counter := 0
	m.mu.Lock()
	m.weightMap = make(map[int]int)
	// count all servers
	// create one or more linsk to all servers
	// quantity of links to one server proportional server weight
	for ind, id := range p {
		s, exists := Servers[id]
		if !exists {
			continue
		}
		s.m.RLock()
		for ii := s.Weight; ii > 0; ii-- {
			m.weightMap[counter] = ind
			counter++
		}
		s.m.RUnlock()
	}
	m.mu.Unlock()
	return p
}

// import (
// 	"fmt"

// 	"github.com/averageNetAdmin/andproxy/proxy/srv"
// )

// // connect to server
// type LeastConnections struct {
// }

// // find server with least connections number and return it (with weight)
// func (m *LeastConnections) FindServer(sIP string, p []int64) (int64, error) {
// 	if len(p) == 0 {
// 		return nil, fmt.Errorf("no server.Servers avaible in pool")
// 	}
// 	srv := p[0]
// 	for i := 1; i < len(p); i++ {
// 		if p[i].GetConnNumber()/uint64(srv.GetWeight()) < srv.GetConnNumber() {
// 			srv = p[i]
// 		}
// 	}
// 	return srv, nil
// }

// // this method is not require rebalancing
// // do nothing
// func (m *LeastConnections) Rebalance(p []int64) {

// }

// no balancing. All requests sends to server with highest priority
// if server down, requsts sends to next server
type None struct {
	mu sync.RWMutex
}

// return first server in pool
func (m *None) FindServer(sIP string, p []int64) (int64, error) {
	if len(p) == 0 {
		return 0, fmt.Errorf("no server.Servers avaible in pool")
	}
	srv := p[0]
	return srv, nil
}

// sort servers by priority
func (m *None) Rebalance(p []int64) []int64 {
	if len(p) == 0 {
		return nil
	}
	weight := int32(0)
	highPriority := 0
	for ind, id := range p {
		s, exists := Servers[id]
		if !exists {
			continue
		}
		s.m.RLock()
		if weight < s.Weight {
			highPriority = ind
			weight = s.Weight
		}
		s.m.RUnlock()
	}
	m.mu.Lock()
	p[0], p[highPriority] = p[highPriority], p[0]
	m.mu.Unlock()
	return p
}

// requsts sends to random server
type Random struct {
	// weight map is requires to balancing with weight
	weightMap map[int]int
	mu        sync.RWMutex
}

// return random server
func (m *Random) FindServer(sIP string, p []int64) (int64, error) {
	if len(p) == 0 {
		return 0, fmt.Errorf("no server.Servers avaible in pool")
	}

	rand.Seed(time.Now().UnixNano())
	m.mu.RLock()
	n := int(rand.Int63()) % len(m.weightMap)
	srv := p[m.weightMap[n]]
	m.mu.RUnlock()
	return srv, nil
}

// update weight map
func (m *Random) Rebalance(p []int64) []int64 {
	if len(p) == 0 {
		return nil
	}
	counter := 0
	m.mu.Lock()
	for ind, id := range p {
		s, exists := Servers[id]
		if !exists {
			continue
		}
		s.m.RLock()
		for ii := s.Weight; ii > 0; ii-- {
			m.weightMap[counter] = ind
			counter++
		}
		s.m.RUnlock()
	}
	m.mu.Unlock()
	return p
}

type RoundRobin struct {
	counter       int
	weightCounter int32
	mu            sync.RWMutex
}

// reutrn next server
func (m *RoundRobin) FindServer(sIP string, p []int64) (int64, error) {
	if len(p) == 0 {
		return 0, fmt.Errorf("no servers avaible in pool")
	}
	m.mu.Lock()
	ind := p[m.counter%len(p)]
	if m.weightCounter >= 1 {
		m.counter++
		ind = p[m.counter%len(p)]
		s, exists := Servers[ind]
		if !exists {
			return 0, fmt.Errorf("server not exists")
		}
		s.m.RLock()
		m.weightCounter = s.Weight
		s.m.RUnlock()
	} else {
		m.weightCounter--
	}
	m.mu.Unlock()
	return ind, nil
}

// update summary weight
func (m *RoundRobin) Rebalance(p []int64) []int64 {
	if len(p) == 0 {
		return nil
	}
	m.mu.Lock()
	s, exists := Servers[p[m.counter%len(p)]]
	if !exists {
		return nil
	}
	s.m.RLock()
	if m.weightCounter > s.Weight {
		m.weightCounter = s.Weight
	}
	s.m.RUnlock()
	m.mu.Unlock()
	return p
}
