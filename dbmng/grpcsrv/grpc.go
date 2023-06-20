package grpcsrv

import (
	"net"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type DbSrv struct {
	andproto.UnimplementedDbmngServer
}

func CreateGRPCServer(port string) error {

	lis, err := net.Listen("tcp", net.JoinHostPort("localhost", port))

	if err != nil {
		err = errors.WithMessage(err, "net.Listen()")
		return err
	}
	// andproto.UnimplementedDbmngServer
	s := grpc.NewServer()
	andproto.RegisterDbmngServer(s, &DbSrv{})
	return s.Serve(lis)
}
