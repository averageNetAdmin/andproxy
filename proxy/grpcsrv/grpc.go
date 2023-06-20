package grpcsrv

import (
	"net"

	"github.com/averageNetAdmin/andproxy/andproto"
	"google.golang.org/grpc"
)

type Proxy struct {
	andproto.UnimplementedProxyServer
}

func CreateGRPCServer(port string) error {

	lis, err := net.Listen("tcp", net.JoinHostPort("localhost", port))

	if err != nil {
		return err
	}
	// andproto.UnimplementedDbmngServer
	s := grpc.NewServer()
	andproto.RegisterProxyServer(s, &Proxy{})
	return s.Serve(lis)
}
