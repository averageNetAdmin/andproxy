package balancing

import (
	"fmt"
)

type LeastConnections struct {
}

func (m *LeastConnections) FindServer(sIP string, p []BalanceItem) (BalanceItem, error) {
	if len(p) == 0 {
		return nil, fmt.Errorf("no server.Servers avaible in pool")
	}
	srv := p[0]
	for i := 1; i < len(p); i++ {
		if p[i].GetConnNumber()/uint64(srv.GetWeight()) < srv.GetConnNumber() {
			srv = p[i]
		}
	}
	return srv, nil
}

func (m *LeastConnections) Rebalance(p []BalanceItem) {

}
