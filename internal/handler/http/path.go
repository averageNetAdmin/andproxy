package http

import (
	"fmt"
	"regexp"
	"time"

	"github.com/averageNetAdmin/andproxy/internal/client"
)

//	Content info about ever site path
//
type Path struct {
	Path                     *regexp.Regexp
	Accept                   *client.Sources
	Deny                     *client.Sources
	Servers                  *Pool
	IPFilter                 []*IPFilter
	Toport                   int
	DeadLine                 time.Duration
	WriteDeadLine            time.Duration
	ReadDeadLine             time.Duration
	MaxConnectTime           time.Duration
	MaxConnections           int64
	connectionsNumber        uint64
	currentconnectionsNumber int64
	rejected                 uint64
	OverFlow                 string
}

func NewPath(URLPath string, config map[string]interface{}) (*Path, error) {

	path, err := regexp.Compile(URLPath)
	if err != nil {
		return nil, err
	}

	accept, err := client.New()
	if err != nil {
		return nil, err
	}
	acceptArr, ok := config["accept"].([]interface{})

	if ok {
		for _, v := range acceptArr {
			acc, ok := v.(string)
			if ok {
				err := accept.Add(acc)
				if err != nil {
					return nil, err
				}
			}

		}
		if len(acceptArr) == 0 {

			accept = nil
		}
	} else {
		accept = nil
	}

	deny, err := client.New()
	if err != nil {
		return nil, err
	}
	denyArr, ok := config["deny"].([]interface{})
	if ok {
		for _, v := range denyArr {
			den, ok := v.(string)
			if ok {
				err := deny.Add(den)
				if err != nil {
					return nil, err
				}
			}
		}
		if len(denyArr) == 0 {
			deny = nil
		}
	} else {
		deny = nil
	}

	srvs := make([]*Server, 0)
	serversArr, ok := config["servers"].([]interface{})
	if ok {
		for i := 0; i < len(serversArr); i++ {
			serversStr, ok := serversArr[i].(map[string]interface{})
			if ok {
				srvss, err := ServersFromMap(serversStr)
				if err != nil {
					return nil, err
				}
				srvs = append(srvs, srvss...)
			}

		}
	}

	balancingStr, ok := config["balancing"].(string)
	if ok {
		balancingStr = ""
	}
	pool, err := NewPool(srvs, balancingStr)
	if err != nil {
		return nil, err
	}

	var (
		dl, rdl, wdl, mconntime time.Duration
		maxconn                 int64
		toport                  int
	)

	if config["deadline"] != nil {
		dlS, ok := config["deadline"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler deadline %v", config["deadline"])
		}
		dl, err = time.ParseDuration(dlS)
		if err != nil {
			return nil, err
		}
	}
	if config["readdeadline"] != nil {
		rdlS, ok := config["readdeadline"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler readdeadline %v", config["readdeadline"])
		}
		rdl, err = time.ParseDuration(rdlS)
		if err != nil {
			return nil, err
		}
	}
	if config["writedeadline"] != nil {
		wdlS, ok := config["writedeadline"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler writedeadline %v", config["writedeadline"])
		}
		wdl, err = time.ParseDuration(wdlS)
		if err != nil {
			return nil, err
		}
	}
	if config["maxconnectionstime"] != nil {
		mconntimeS, ok := config["maxconnectionstime"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid handler maxconnectionstime %v", config["maxconnectionstime"])
		}
		mconntime, err = time.ParseDuration(mconntimeS)
		if err != nil {
			return nil, err
		}
	}
	if config["maxconnections"] != nil {
		mc, ok := config["maxconnections"].(int)
		if !ok {
			return nil, fmt.Errorf("invalid server maxconnections %v", config["maxconnections"])
		}
		maxconn = int64(mc)
	}
	if config["toport"] != nil {
		toport, ok = config["toport"].(int)
		if !ok {
			return nil, fmt.Errorf("invalid handler destiantion port %v", config["toport"])
		}
	}

	filters := make([]*IPFilter, 0)
	filtersStr, ok := config["ipfilters"].([]map[string]interface{})
	if ok {
		for i := 0; i < len(filtersStr); i++ {
			srvs := make([]*Server, 0)
			serversStr, ok := filtersStr[i]["servers"].([]map[string]interface{})
			if ok {
				for i := 0; i < len(serversStr); i++ {
					srvss, err := ServersFromMap(serversStr[i])
					if err != nil {
						return nil, err
					}
					srvs = append(srvs, srvss...)
				}
			}
			balancingStr, ok := filtersStr[i]["balancing"].(string)
			if ok {
				balancingStr = ""
			}
			pool, err := NewPool(srvs, balancingStr)
			if err != nil {
				return nil, err
			}

			var source *client.Sources
			sourceStr, ok := filtersStr[i]["source"].([]string)
			if ok {
				source, err = client.New(sourceStr...)
				if err != nil {
					return nil, err
				}
			}
			filter := NewFilter(pool, source)
			filters = append(filters, filter)
		}

	}

	return &Path{
		Path:           path,
		Toport:         toport,
		Accept:         accept,
		Deny:           deny,
		Servers:        pool,
		IPFilter:       filters,
		DeadLine:       dl,
		WriteDeadLine:  wdl,
		ReadDeadLine:   rdl,
		MaxConnectTime: mconntime,
		MaxConnections: maxconn,
	}, err

}
