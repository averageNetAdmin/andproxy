package balancing

import (
	"fmt"
	"hash/fnv"
	"sync"
)

// Filter requests by client ip address hash
// Client always go to same server
//
type HashIP struct {
	weightMap map[int]int
	mu        sync.RWMutex
}

func (m *HashIP) FindServer(sIP string, p []BalanceItem) (BalanceItem, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no server.Servers avaible in pool")
	}
	h := fnv.New32a()
	h.Write([]byte(sIP))
	m.mu.RLock()
	srv := p[m.weightMap[int(h.Sum32())%len(m.weightMap)]]
	m.mu.RUnlock()
	return srv, nil
}

// If count of servers was changed, weight map must be changed
// 
func (m *HashIP) Rebalance(p []BalanceItem) {
	if len(p) == 0 {
		return
	}
	counter := 0
	m.mu.Lock()
	// count all servers
	// create one or more linsk to all servers
	// quantity of links to one server proportional server weight
	for i := 0; i < len(p); i++ {
		for ii := p[i].GetWeight(); ii > 0; ii-- {
			m.weightMap[counter] = i
			counter++
		}
	}
	m.mu.Unlock()
}
