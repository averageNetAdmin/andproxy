package balancing

import (
	"fmt"
)

type None struct {
}

func (m *None) FindServer(sIP string, p []BalanceItem) (BalanceItem, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no server.Servers avaible in pool")
	}
	srv := p[0]
	return srv, nil
}

func (m *None) Rebalance(p []BalanceItem) {
	if len(p) == 0 {
		return
	}
	weight := 0
	highPriority := 0
	for i := 0; i < len(p); i++ {
		if weight < p[i].GetWeight() {
			highPriority = i
			weight = p[i].GetWeight()
		}
	}
	p[0], p[highPriority] = p[highPriority], p[0]
}
