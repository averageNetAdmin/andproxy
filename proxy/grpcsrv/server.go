package grpcsrv

import (
	"context"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/srv"
	"google.golang.org/protobuf/types/known/emptypb"
)

func InitServerCashe() error {

	resp, err := config.DbMngClient.ReadServers(context.Background(), &andproto.ReadServersRequest{})
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadServers()")
		return err
	}

	for _, item := range resp.Srv {
		srv.Servers[item.Id] = &srv.Server{
			Server: item,
		}
	}

	return nil
}

func (d *Proxy) CreateServer(ctx context.Context, request *andproto.CreateServerRequest) (*andproto.CreateServerResponse, error) {

	funcName := "CreateServer"

	response, err := config.DbMngClient.CreateServer(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.CreateServer()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	request.Srv.Id = response.Id

	srv.ServersMutex.Lock()
	srv.Servers[response.Id] = &srv.Server{
		Server: request.Srv,
	}
	srv.ServersMutex.Unlock()

	return response, nil
}

func (d *Proxy) UpdateServer(ctx context.Context, request *andproto.UpdateServerRequest) (*emptypb.Empty, error) {

	funcName := "UpdateServer"

	_, err := config.DbMngClient.UpdateServer(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.UpdateServer()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	srv.ServersMutex.Lock()
	srv.Servers[request.Srv.Id] = &srv.Server{
		Server: request.Srv,
	}
	srv.ServersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) DeleteServer(ctx context.Context, request *andproto.DeleteServerRequest) (*emptypb.Empty, error) {

	funcName := "DeleteServer"

	_, err := config.DbMngClient.DeleteServer(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.DeleteServer()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	srv.ServersMutex.Lock()
	delete(srv.Servers, request.Id)
	srv.ServersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) ReadServers(ctx context.Context, request *andproto.ReadServersRequest) (*andproto.ReadServersResponse, error) {

	funcName := "ReadServers"

	resp, err := config.DbMngClient.ReadServers(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadServers()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}
	return resp, nil
}
