package httph

import (
	"crypto/tls"
	"regexp"
	"sync"

	"github.com/averageNetAdmin/andproxy/andproto/models"
)

var HttpHostsMutex = &sync.RWMutex{}
var HttpHosts = make(map[int64]*HttpHost)

type HttpHost struct {
	*models.HttpHost
	HostReg     *regexp.Regexp
	m           sync.RWMutex
	Certificate *tls.Certificate
}
