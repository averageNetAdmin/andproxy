package balancing

import (
	"net"

	"github.com/averageNetAdmin/andproxy/source/ippool"
)

type Method interface {
	Find(ip string, p *ippool.ServerPool) (string, error)
}

type RoundRobin struct {
	counter int
}

func (m RoundRobin) Find(sIP string, p *ippool.ServerPool, proto, port string) (net.Conn, error) {
	for i := len(p.Servers);
	net.Dial(proto, ip + ":"+port)
}
