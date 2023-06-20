package srv

import (
	"fmt"
	"net"
	"sync"

	"github.com/averageNetAdmin/andproxy/andproto/models"
	"github.com/averageNetAdmin/andproxy/proxy/ac"
)

var ProxyFiltersMutex = &sync.RWMutex{}
var ProxyFilters = make(map[int64]*ProxyFilter)

type ProxyFilter struct {
	*models.ProxyFilter
	TargetNet *net.IPNet
	Accept    *ac.Acl
	Deny      *ac.Acl
	m         sync.RWMutex
}

func (f *ProxyFilter) GetPool(ip string) (int64, error) {

	reqIP := net.ParseIP(ip)
	f.m.RLock()
	defer f.m.RUnlock()
	if !f.TargetNet.Contains(reqIP) {
		return 0, fmt.Errorf("access denied")
	}
	if f.Accept != nil && f.Accept.Contains(ip) {
		return f.ServerPool, nil
	}
	if f.Deny != nil && f.Deny.Contains(ip) {
		return 0, fmt.Errorf("access denied")
	}

	return f.ServerPool, nil
}
