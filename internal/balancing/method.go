package balancing

import (
	"fmt"
)

// interface to object that can be balanced
//
type BalanceItem interface {
	GetWeight() int
	GetConnNumber() uint64
}

// Interface to balancing method
//
type Method interface {
	Rebalance([]BalanceItem)
	FindServer(string, []BalanceItem) (BalanceItem, error)
}

// return new balancing method with checked name
//
func NewMethod(name string) (Method, error) {
	switch name {
	case "roundrobin":
		return &RoundRobin{counter: 0, weightCounter: 1}, nil
	case "none":
		return &None{}, nil
	case "random":
		return &Random{weightMap: make(map[int]int)}, nil
	case "haship":
		return &HashIP{weightMap: make(map[int]int)}, nil
	case "leastconnections":
		return &LeastConnections{}, nil
	default:
		return nil, fmt.Errorf("%s balancing method not exist", name)
	}
}
