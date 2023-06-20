package handler

import "github.com/averageNetAdmin/andproxy/proxy/errs"

type Handler interface {
	StartHandle()
	StopHandle()
}

func New(proto, port string, config []byte) (*Handler, error) {
	if proto == "" {
		return nil, errs.ErrEmptyHandlerProtocol
	}
	if proto == "" {
		return nil, errs.ErrEmptyHandlerProtocol
	}
	switch proto {
	case "tcp", "udp":
	default:
		errs.NewInvlidHandlerProtocolErr(proto, port)
	}
	return nil, nil
}
