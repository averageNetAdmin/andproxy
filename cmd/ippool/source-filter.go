package ippool

//
//	the struct that Contains array of filgerElements
//
type Filter struct {
	filters []filterElement
}

//
//	the struct that compare cliets addresses and servers to which should be forwarded requests
//
type filterElement struct {
	Pool    Pool
	Srvpool ServerPool
}

//
//	chcek IP pool contain IP address or not
//
func (elem *filterElement) Contains(ip string) (bool, error) {
	yes, err := elem.Pool.Contains(ip)
	if err != nil {
		return false, err
	}
	return yes, nil
}

func (elem *filterElement) Rebalance() {
	elem.Srvpool.Rebalance()
}

func (f *Filter) Rebalance() {
	for i := 0; i < len(f.filters); i++ {
		f.filters[i].Srvpool.Rebalance()
	}
}

func (f *Filter) SetLogFile(logDir string) error {
	for i := 0; i < len(f.filters); i++ {
		err := f.filters[i].Srvpool.SetLogFile(logDir)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *Filter) SetBalancingMethod(bm string) error {
	for i := 0; i < len(f.filters); i++ {
		err := f.filters[i].Srvpool.SetBalancingMethod(bm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *filterElement) SetBalancingMethod(bm string) error {
	err := f.Srvpool.SetBalancingMethod(bm)
	return err
}

//
//	i don't know that to say
//
func (elem *filterElement) SetPool(p *Pool) {
	elem.Pool = *p
}

//
//	i don't know that to say
//
func (elem *filterElement) SetServerPool(p *ServerPool) {
	elem.Srvpool = *p
}

//
//	i don't know that to say
//
func (elem *filterElement) Set(pool *Pool, srvpool *ServerPool) {
	elem.Pool = *pool
	elem.Srvpool = *srvpool
}

//
//	create and return new filterElement
//
func newFilterElement(pool *Pool, srvpool *ServerPool) filterElement {
	return filterElement{
		Pool:    *pool,
		Srvpool: *srvpool,
	}
}

//
//	i don't know that to say
//
func (f *Filter) Add(pool *Pool, srvpool *ServerPool) {
	elem := newFilterElement(pool, srvpool)
	f.filters = append(f.filters, elem)
}

//
//	search what pool should handle request
//
func (f *Filter) WhatPool(ip string) (*ServerPool, error) {
	for _, felem := range f.filters {
		isContent, err := felem.Contains(ip)
		if err != nil {
			return nil, err
		}
		if isContent {
			return &felem.Srvpool, nil
		}
	}

	return nil, nil
}
