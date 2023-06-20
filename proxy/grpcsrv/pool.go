package grpcsrv

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/srv"
	"google.golang.org/protobuf/types/known/emptypb"
)

func InitSrvPoolCashe() error {

	resp, err := config.DbMngClient.ReadServerPools(context.Background(), &andproto.ReadServerPoolsRequest{})
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadServerPools()")
		return err
	}

	for _, item := range resp.Pool {

		bm, err := srv.NewMethod(item.BalancingMethod)
		if err != nil {
			err = errors.WithMessage(err, "srv.NewMethod(()")
			return err
		}
		srv.ServerPools[item.Id] = &srv.ServerPool{
			ServerPool: item,
			Active:     make([]int64, len(item.Servers)),
			BM:         bm,
		}
		copy(srv.ServerPools[item.Id].Active, item.Servers)
	}

	return nil
}

func (d *Proxy) CreateServerPool(ctx context.Context, request *andproto.CreateServerPoolRequest) (*andproto.CreateServerPoolResponse, error) {

	funcName := "CreateServerPool"

	bm, err := srv.NewMethod(request.Pool.BalancingMethod)
	if err != nil {
		err = errors.WithMessage(err, "srv.NewMethod(()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	response, err := config.DbMngClient.CreateServerPool(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.CreateServerPool()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}
	request.Pool.Id = response.Id

	srv.ServerPoolsMutex.Lock()
	srv.ServerPools[response.Id] = &srv.ServerPool{
		ServerPool: request.Pool,
		Active:     make([]int64, len(request.Pool.Servers)),
		BM:         bm,
	}
	copy(srv.ServerPools[response.Id].Active, request.Pool.Servers)
	fmt.Println(srv.ServerPools[response.Id].Active, request.Pool.Servers)

	srv.ServerPoolsMutex.Unlock()

	return response, nil
}

func (d *Proxy) UpdateServerPool(ctx context.Context, request *andproto.UpdateServerPoolRequest) (*emptypb.Empty, error) {

	funcName := "UpdateServerPool"

	bm, err := srv.NewMethod(request.Pool.BalancingMethod)
	if err != nil {
		err = errors.WithMessage(err, "srv.NewMethod(()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	_, err = config.DbMngClient.UpdateServerPool(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.UpdateServerPool()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	srv.ServerPoolsMutex.Lock()
	srv.ServerPools[request.Pool.Id] = &srv.ServerPool{
		ServerPool: request.Pool,
		Active:     make([]int64, len(request.Pool.Servers)),
		BM:         bm,
	}
	copy(srv.ServerPools[request.Pool.Id].Active, request.Pool.Servers)
	srv.ServerPoolsMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) DeleteServerPool(ctx context.Context, request *andproto.DeleteServerPoolRequest) (*emptypb.Empty, error) {

	funcName := "DeleteServerPool"

	_, err := config.DbMngClient.DeleteServerPool(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.DeleteServerPool()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	srv.ServerPoolsMutex.Lock()
	delete(srv.ServerPools, request.Id)
	srv.ServerPoolsMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) ReadServerPools(ctx context.Context, request *andproto.ReadServerPoolsRequest) (*andproto.ReadServerPoolsResponse, error) {

	funcName := "ReadServerPools"

	resp, err := config.DbMngClient.ReadServerPools(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadServerPools()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}
	return resp, nil
}
