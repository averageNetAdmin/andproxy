package balancing

import (
	"fmt"
	"sync"
)

type RoundRobin struct {
	counter       int
	weightCounter int
	mu            sync.RWMutex
}

// reutrn next server
//
func (m *RoundRobin) FindServer(sIP string, p []BalanceItem) (BalanceItem, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no servers avaible in pool")
	}
	var srv BalanceItem
	m.mu.Lock()
	if m.weightCounter >= 1 {
		m.counter++
		srv = p[m.counter%len(p)]
		m.weightCounter = srv.GetWeight()
	} else {
		srv = p[m.counter%len(p)]
		m.weightCounter--
	}
	m.mu.Unlock()
	return srv, nil
}

// update summary weight
//
func (m *RoundRobin) Rebalance(p []BalanceItem) {
	if len(p) == 0 {
		return
	}
	m.mu.Lock()
	if m.weightCounter > p[m.counter%len(p)].GetWeight() {
		m.weightCounter = p[m.counter%len(p)].GetWeight()
	}
	m.mu.Unlock()
}
