package srcfltr

import (
	"github.com/averageNetAdmin/andproxy/source/ippool"
)

//
//	the struct that contains array of filgerElements
//
type Filter struct {
	filters []filterElement
}

//
//	the struct that compare cliets addresses and servers to which should be forwarded requests
//
type filterElement struct {
	pool    *ippool.Pool
	srvpool *ippool.ServerPool
}

//
//	chcek IP pool contain IP address or not
//
func (elem *filterElement) Contains(ip string) (bool, error) {
	yes, err := elem.pool.Conatains(ip)
	if err != nil {
		return false, err
	}
	return yes, nil
}

//
//	i don't know that to say
//
func (elem *filterElement) SetPool(p *ippool.Pool) {
	elem.pool = p
}

//
//	i don't know that to say
//
func (elem *filterElement) SetServerPool(p *ippool.ServerPool) {
	elem.srvpool = p
}

//
//	i don't know that to say
//
func (elem *filterElement) Set(pool *ippool.Pool, srvpool *ippool.ServerPool) {
	elem.pool = pool
	elem.srvpool = srvpool
}

//
//	create and return new filterElement
//
func newFilterElement(pool *ippool.Pool, srvpool *ippool.ServerPool) filterElement {
	return filterElement{
		pool:    pool,
		srvpool: srvpool,
	}
}

//
//	i don't know that to say
//
func (f *Filter) Add(pool *ippool.Pool, srvpool *ippool.ServerPool) {
	elem := newFilterElement(pool, srvpool)
	f.filters = append(f.filters, elem)
}

//
//	show filter content
//
func (f *Filter) String() string {
	result := ""
	for _, felem := range f.filters {
		result += felem.pool.String() + " "
		result += felem.srvpool.String() + "\n"
	}
	return result
}

//
//	search what pool should handle request
//
func (f *Filter) WhatPool(ip string) (*ippool.ServerPool, error) {
	for _, felem := range f.filters {
		isContent, err := felem.Contains(ip)
		if err != nil {
			return nil, err
		}
		if isContent {
			return felem.srvpool, nil
		}
	}

	return nil, nil
}
