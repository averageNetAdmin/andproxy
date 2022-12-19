package balancing

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// requsts sends to random server
//
type Random struct {
	// weight map is requires to balancing with weight 
	weightMap map[int]int
	mu        sync.RWMutex
}

// return random server
//
func (m *Random) FindServer(sIP string, p []BalanceItem) (BalanceItem, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no server.Servers avaible in pool")
	}

	rand.Seed(time.Now().UnixNano())
	m.mu.RLock()
	n := int(rand.Int63()) % len(m.weightMap)
	srv := p[m.weightMap[n]]
	m.mu.RUnlock()
	return srv, nil
}

// update weight map
//
func (m *Random) Rebalance(p []BalanceItem) {
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
