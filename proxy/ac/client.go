package ac

import (
	"net"
	"sync"

	"github.com/averageNetAdmin/andproxy/andproto/models"
)

var AclsMutex = &sync.RWMutex{}
var Acls = make(map[int64]*Acl)

type Acl struct {
	*models.Acl
	Nets []*net.IPNet
}

func (acl *Acl) Contains(ip string) bool {

	sip := net.ParseIP(ip)

	if acl.Nets != nil {
		for _, ace := range acl.Nets {
			if ace.Contains(sip) {
				return true
			}
		}
	}

	return false
}
