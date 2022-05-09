package ipdb

import (
	"fmt"
	"strings"

	"github.com/averageNetAdmin/andproxy/cmd/ippool"
)

//
//	struct what contain all pools and filters in programm
//
type IPDB struct {
	pools       map[string]*ippool.Pool
	serverPools map[string]*ippool.ServerPool
	filters     map[string]*ippool.Filter
}

//
//	return new ip database
//
func NewIPDB() *IPDB {
	db := new(IPDB)
	db.Init()
	return db
}

//
//	init db maps
//
func (db *IPDB) Init() {
	db.pools = make(map[string]*ippool.Pool)
	db.serverPools = make(map[string]*ippool.ServerPool)
	db.filters = make(map[string]*ippool.Filter)
}

//
//	add pool to db
//
func (db *IPDB) AddPool(name string, addresses []interface{}) error {
	addrs := make([]string, 0)
	for _, a := range addresses {
		if a == nil {
			break
		}
		addrs = append(addrs, a.(string))
	}
	p, err := ippool.NewPool(addrs...)
	if err != nil {
		return err
	}
	p.Name = name
	db.pools[name] = p
	return nil
}

//
//	add servers pool to db
//
func (db *IPDB) AddServerPool(name string, addresses map[string]interface{}) error {
	p, err := ippool.NewServerPool(addresses)
	if err != nil {
		return err
	}
	p.Name = name
	db.serverPools[name] = p
	return nil
}

//
//	add filter to db
//
func (db *IPDB) AddFilter(name string, elements map[string]interface{}) error {
	result := new(ippool.Filter)
	for pool, srvPool := range elements {
		if srvPool == nil {
			continue
		}

		var resPool *ippool.Pool
		var resServerPool *ippool.ServerPool
		isVar := strings.HasPrefix(pool, "$")
		if isVar {
			p := db.pools[pool[1:]]
			if p == nil {
				return fmt.Errorf("error: pool %v is not exist in filter %s", pool, name)
			}
			resPool = p
		} else {
			p, err := ippool.NewPool(pool)
			if err != nil {
				return err
			}
			resPool = p
		}

		if sPool, ok := srvPool.(string); ok && strings.HasPrefix(sPool, "$") {
			p := db.serverPools[sPool[1:]]
			if p == nil {
				return fmt.Errorf("error: server pool name %v is not exist in filter %s", sPool, name)
			}
			resServerPool = p
		} else if sPool, ok := srvPool.(map[string]interface{}); ok {
			p, err := ippool.NewServerPool(sPool)
			if err != nil {
				return err
			}
			resServerPool = p
		} else {
			return fmt.Errorf("error: invalid server pool %v in filter %s", srvPool, name)
		}
		result.Add(resPool, resServerPool)
	}
	result.Name = name
	db.filters[name] = result
	return nil
}

//
//	...
//
func (db *IPDB) GetPool(name string) *ippool.Pool {
	return db.pools[name]
}

//
//	...
//
func (db *IPDB) GetServerPool(name string) *ippool.ServerPool {
	return db.serverPools[name]
}

//
//	...
//
func (db *IPDB) GetFilter(name string) *ippool.Filter {
	return db.filters[name]
}

//
//	...
//
func (db *IPDB) GetPoolCopy(name string) (ippool.Pool, bool) {
	res, ok := db.pools[name]
	if !ok {
		return ippool.Pool{}, false
	}
	return *res, true
}

//
//	...
//
func (db *IPDB) GetServerPoolCopy(name string) (ippool.ServerPool, bool) {
	res, ok := db.serverPools[name]
	if !ok {
		return ippool.ServerPool{}, false
	}
	return *res, true
}

//
//	...
//
func (db *IPDB) GetFilterCopy(name string) (ippool.Filter, bool) {
	res, ok := db.filters[name]
	if !ok {
		return ippool.Filter{}, false
	}
	return *res, true
}
