package grpcsrv

import (
	"context"
	"regexp"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/handler/httph"
	"google.golang.org/protobuf/types/known/emptypb"
)

func InitHttpPathCashe() error {

	resp, err := config.DbMngClient.ReadHttpPaths(context.Background(), &andproto.ReadHttpPathsRequest{})
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadHttpPaths()")
		return err
	}

	for _, item := range resp.Path {

		reg, err := regexp.Compile(item.Path)
		if err != nil {
			err = errors.WithMessage(err, "regexp.Compile()")
			return err
		}

		httph.HttpPaths[item.Id] = &httph.HttpPath{
			HttpPath: item,
			PathReg:  reg,
		}
	}

	return nil
}

func (d *Proxy) CreateHttpPath(ctx context.Context, request *andproto.CreateHttpPathRequest) (*andproto.CreateHttpPathResponse, error) {

	funcName := "CreateHttpPath"

	reg, err := regexp.Compile(request.Path.Path)
	if err != nil {
		err = errors.WithMessage(err, "regexp.Compile()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	response, err := config.DbMngClient.CreateHttpPath(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.CreateHttpPath()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	request.Path.Id = response.Id

	httph.HttpPathsMutex.Lock()
	httph.HttpPaths[response.Id] = &httph.HttpPath{
		HttpPath: request.Path,
		PathReg:  reg,
	}
	httph.HttpPathsMutex.Unlock()

	return response, nil
}

func (d *Proxy) UpdateHttpPath(ctx context.Context, request *andproto.UpdateHttpPathRequest) (*emptypb.Empty, error) {

	funcName := "UpdateHttpPath"

	reg, err := regexp.Compile(request.Path.Path)
	if err != nil {
		err = errors.WithMessage(err, "regexp.Compile()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	_, err = config.DbMngClient.UpdateHttpPath(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.UpdateHttpPath()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	httph.HttpPathsMutex.Lock()
	httph.HttpPaths[request.Path.Id] = &httph.HttpPath{
		HttpPath: request.Path,
		PathReg:  reg,
	}
	httph.HttpPathsMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) DeleteHttpPath(ctx context.Context, request *andproto.DeleteHttpPathRequest) (*emptypb.Empty, error) {

	funcName := "DeleteHttpPath"

	_, err := config.DbMngClient.DeleteHttpPath(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.DeleteHttpPath()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	httph.HttpPathsMutex.Lock()
	delete(httph.HttpPaths, request.Id)
	httph.HttpPathsMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) ReadHttpPaths(ctx context.Context, request *andproto.ReadHttpPathsRequest) (*andproto.ReadHttpPathsResponse, error) {

	funcName := "ReadHttpPaths"

	resp, err := config.DbMngClient.ReadHttpPaths(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadHttpPaths()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}
	return resp, nil
}
