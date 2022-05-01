package ipdb

import (
	"errors"
	"strings"

	"github.com/averageNetAdmin/andproxy/bin/ippool"
	"github.com/averageNetAdmin/andproxy/bin/srcfltr"
)

//
//	struct what contain all pools and filters in programm
//
type IPDB struct {
	pools       map[string]*ippool.Pool
	serverPools map[string]*ippool.ServerPool
	filters     map[string]*srcfltr.Filter
}

//
//	return new ip database
//
func NewIPDB() (*IPDB) {
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
	db.filters = make(map[string]*srcfltr.Filter)
}

//
//	add pool to db
//
func (db *IPDB) AddPool(name string, addresses []interface{}) error {
	db.pools = make(map[string]*ippool.Pool)
	db.serverPools = make(map[string]*ippool.ServerPool)
	db.filters = make(map[string]*srcfltr.Filter)
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
	db.pools[name] = p
	return nil
}

//
//	add servers pool to db
//
func (db *IPDB) AddServerPool(name string, addresses []interface{}) error {
	addrs := make([]string, 0)
	for _, a := range addresses {
		if a == nil {
			break
		}
		addrs = append(addrs, a.(string))
	}
	p, err := ippool.NewServerPool(addrs...)
	if err != nil {
		return err
	}
	db.serverPools[name] = p
	return nil
}

//
//	add filter to db
//
func (db *IPDB) AddFilter(name string, elements map[string]interface{}) error {
	result := new(srcfltr.Filter)
	for pool, srvPool := range elements {
		if srvPool == nil {
			continue
		}
		sPool := srvPool.(string)
		var resPool *ippool.Pool
		var resServerPool *ippool.ServerPool
		isVar := strings.HasPrefix(pool, "$")
		if isVar {
			p := db.pools[pool]
			if p == nil {
				return errors.New("Pool" + pool + " is not exist")
			}
			resPool = p
		} else {
			p, err := ippool.NewPool(pool)
			if err != nil {
				return err
			}
			resPool = p
		}
		isVar = strings.HasPrefix(sPool, "$")
		if isVar {
			p := db.serverPools[sPool]
			if p == nil {
				return errors.New("Pool" + sPool + " is not exist")
			}
			resServerPool = p
		} else {
			p, err := ippool.NewServerPool(sPool)
			if err != nil {
				return err
			}
			resServerPool = p
		}
		result.Add(resPool, resServerPool)
	}
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
func (db *IPDB) GetFilter(name string) *srcfltr.Filter {
	return db.filters[name]
}
