package grpcsrv

import (
	"context"
	"net"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/srv"
	"google.golang.org/protobuf/types/known/emptypb"
)

func InitFilterCashe() error {

	resp, err := config.DbMngClient.ReadProxyFilters(context.Background(), &andproto.ReadProxyFiltersRequest{})
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadProxyFilters()")
		return err
	}

	for _, item := range resp.Fltr {

		srv.ProxyFilters[item.Id] = &srv.ProxyFilter{
			ProxyFilter: item,
		}
		_, srv.ProxyFilters[item.Id].TargetNet, err = net.ParseCIDR(item.TargetNet)
		if err != nil {
			err = errors.WithMessage(err, "net.ParseCIDR()")
			return err
		}

	}

	return nil
}

func (d *Proxy) CreateProxyFilter(ctx context.Context, request *andproto.CreateProxyFilterRequest) (*andproto.CreateProxyFilterResponse, error) {

	funcName := "CreateProxyFilter"

	_, nt, err := net.ParseCIDR(request.Fltr.TargetNet)
	if err != nil {
		err = errors.WithMessage(err, "net.ParseCIDR()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	response, err := config.DbMngClient.CreateProxyFilter(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.CreateProxyFilter()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	request.Fltr.Id = response.Id

	srv.ProxyFiltersMutex.Lock()
	srv.ProxyFilters[response.Id] = &srv.ProxyFilter{
		ProxyFilter: request.Fltr,
		TargetNet:   nt,
	}
	srv.ProxyFiltersMutex.Unlock()

	return response, nil
}

func (d *Proxy) UpdateProxyFilter(ctx context.Context, request *andproto.UpdateProxyFilterRequest) (*emptypb.Empty, error) {

	funcName := "UpdateProxyFilter"

	_, nt, err := net.ParseCIDR(request.Fltr.TargetNet)
	if err != nil {
		err = errors.WithMessage(err, "net.ParseCIDR()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	_, err = config.DbMngClient.UpdateProxyFilter(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.UpdateProxyFilter()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	srv.ProxyFiltersMutex.Lock()
	srv.ProxyFilters[request.Fltr.Id] = &srv.ProxyFilter{
		ProxyFilter: request.Fltr,
		TargetNet:   nt,
	}
	srv.ProxyFiltersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) DeleteProxyFilter(ctx context.Context, request *andproto.DeleteProxyFilterRequest) (*emptypb.Empty, error) {

	funcName := "DeleteProxyFilter"

	_, err := config.DbMngClient.DeleteProxyFilter(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.DeleteProxyFilter()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	srv.ProxyFiltersMutex.Lock()
	delete(srv.ProxyFilters, request.Id)
	srv.ProxyFiltersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) ReadProxyFilters(ctx context.Context, request *andproto.ReadProxyFiltersRequest) (*andproto.ReadProxyFiltersResponse, error) {

	funcName := "ReadProxyFilters"

	resp, err := config.DbMngClient.ReadProxyFilters(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadProxyFilters()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}
	return resp, nil
}
