package def

import (
	"github.com/averageNetAdmin/andproxy/internal/client"
)

//	Compare servers ip addresses with source addresses
//
type IPFilter struct {
	source  *client.Sources
	servers *Pool
}

//	Check is ip address in struct
//	Return true if struct contains searchIP ip address else return false
//	If searchIP is not valid ip address return false
//
func (f *IPFilter) Contains(ip string) *Pool {
	content := f.source.Contains(ip)
	if content {
		return f.servers
	}
	return nil
}

// Createt new Filter from servers Pool and Sources
//
func NewFilter(pool *Pool, from *client.Sources) *IPFilter {
	return &IPFilter{
		servers: pool,
		source:  from,
	}
}
