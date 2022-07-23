package def

import (
	"github.com/averageNetAdmin/andproxy/internal/client"
)

type IPFilter struct {
	source  *client.Sources
	servers *Pool
}

func (f *IPFilter) Contains(ip string) *Pool {
	content := f.source.Contains(ip)
	if content {
		return f.servers
	}
	return nil
}

func NewFilter(pool *Pool, from *client.Sources) *IPFilter {
	return &IPFilter{
		servers: pool,
		source:  from,
	}
}
