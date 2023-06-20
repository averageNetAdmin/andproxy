package srv

import (
	"sync"
	"time"

	"github.com/averageNetAdmin/andproxy/andproto/models"
)

var ServerPoolsMutex = &sync.RWMutex{}
var ServerPools = make(map[int64]*ServerPool)

type ServerPool struct {
	*models.ServerPool
	BM     BalancingMethod
	m      sync.RWMutex
	Active []int64
}

func (p *ServerPool) GetSrv(sourceIp string) int64 {

	for i := 0; i < len(p.Active); i++ {

		sind, err := p.BM.FindServer(sourceIp, p.Active)
		if err != nil {
			return 0
		}
		s, exists := Servers[sind]
		if !exists {
			continue
		}
		success := s.TryConnect()
		if success {
			return sind
		}
		if s.IsDown() {
			p.Deactivate(sind)
			time.AfterFunc(time.Duration(s.BreakTime), func() {
				id := sind
				p.Activate(id)
			})
		}
	}
	return 0
}

func (p *ServerPool) Deactivate(id int64) {
	p.m.Lock()
	defer p.m.Unlock()
	for i, v := range p.Active {
		if v == id {
			p.Active = append(p.Active[:i], p.Active[i+1:]...)
			return
		}

	}

}

func (p *ServerPool) Activate(id int64) {
	p.m.Lock()
	defer p.m.Unlock()
	for _, v := range p.Servers {
		if v == id {
			s, exists := Servers[id]
			if !exists {
				continue
			}
			s.m.Lock()
			s.Up()
			s.m.Unlock()
			p.Active = append(p.Active, v)
			return
		}
	}
}
