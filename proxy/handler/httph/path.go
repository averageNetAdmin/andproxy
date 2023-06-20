package httph

import (
	"regexp"
	"sync"

	"github.com/averageNetAdmin/andproxy/andproto/models"
)

var HttpPathsMutex = &sync.RWMutex{}
var HttpPaths = make(map[int64]*HttpPath)

type HttpPath struct {
	*models.HttpPath
	PathReg *regexp.Regexp
	m       sync.RWMutex
}
