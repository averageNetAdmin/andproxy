package balancing

import (
	"fmt"
	"hash/fnv"
	"sync"
)

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

func (m *HashIP) Rebalance(p []BalanceItem) {
	if len(p) == 0 {
		return
	}
	counter := 0
	m.mu.Lock()
	for i := 0; i < len(p); i++ {
		for ii := p[i].GetWeight(); ii > 0; ii-- {
			m.weightMap[counter] = i
			counter++
		}
	}
	m.mu.Unlock()
}
