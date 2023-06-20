package grpcsrv

import (
	"context"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/andproto/models"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/handler/httph"
	"google.golang.org/protobuf/types/known/emptypb"
)

func InitHttpHandlerCashe() error {

	resp, err := config.DbMngClient.ReadHttpHandlers(context.Background(), &andproto.ReadHttpHandlersRequest{})
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadHttpHandlers()")
		return err
	}

	for _, item := range resp.Hndlr {

		hndlr := &httph.HttpHandler{
			HttpHandler: item,
		}

		if hndlr.Status == models.HttpHandler_Listen {
			hndlr.Listen()
		}

		httph.HttpHandlers[item.Id] = hndlr
	}

	return nil
}

func (d *Proxy) CreateHttpHandler(ctx context.Context, request *andproto.CreateHttpHandlerRequest) (*andproto.CreateHttpHandlerResponse, error) {

	funcName := "CreateHttpHandler"

	response, err := config.DbMngClient.CreateHttpHandler(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.CreateHttpHandler()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	request.Hndlr.Id = response.Id

	httph.HttpHandlersMutex.Lock()
	httph.HttpHandlers[response.Id] = nil
	httph.HttpHandlersMutex.Unlock()

	hndlr := &httph.HttpHandler{
		HttpHandler: request.Hndlr,
	}

	if hndlr.Status == models.HttpHandler_Listen {
		hndlr.Listen()
	}

	httph.HttpHandlersMutex.Lock()
	httph.HttpHandlers[response.Id] = hndlr
	httph.HttpHandlersMutex.Unlock()

	return response, nil
}

func (d *Proxy) UpdateHttpHandler(ctx context.Context, request *andproto.UpdateHttpHandlerRequest) (*emptypb.Empty, error) {

	funcName := "UpdateHttpHandler"

	_, err := config.DbMngClient.UpdateHttpHandler(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.UpdateHttpHandler()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	hndlr := &httph.HttpHandler{
		HttpHandler: request.Hndlr,
	}

	if hndlr.Status == models.HttpHandler_Listen {
		hndlr.Listen()
	}

	httph.HttpHandlersMutex.Lock()
	httph.HttpHandlers[request.Hndlr.Id] = hndlr
	httph.HttpHandlersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) DeleteHttpHandler(ctx context.Context, request *andproto.DeleteHttpHandlerRequest) (*emptypb.Empty, error) {

	funcName := "DeleteHttpHandler"

	_, err := config.DbMngClient.DeleteHttpHandler(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.DeleteHttpHandler()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	httph.HttpHandlersMutex.Lock()
	delete(httph.HttpHandlers, request.Id)
	httph.HttpHandlersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) ReadHttpHandlers(ctx context.Context, request *andproto.ReadHttpHandlersRequest) (*andproto.ReadHttpHandlersResponse, error) {

	funcName := "ReadHttpHandlers"

	resp, err := config.DbMngClient.ReadHttpHandlers(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadHttpHandlers()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}
	return resp, nil
}
