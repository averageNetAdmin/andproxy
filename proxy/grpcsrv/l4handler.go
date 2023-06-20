package grpcsrv

import (
	"context"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/andproto/models"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/handler/l4handler"
	"google.golang.org/protobuf/types/known/emptypb"
)

func InitL4HandlerCashe() error {

	resp, err := config.DbMngClient.ReadL4Handlers(context.Background(), &andproto.ReadL4HandlersRequest{})
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadL4Handlers()")
		return err
	}

	for _, item := range resp.Hndlr {
		l4handler.L4Handlers[item.Id] = &l4handler.L4Handler{
			L4Handler: item,
		}
	}

	return nil
}

func (d *Proxy) CreateL4Handler(ctx context.Context, request *andproto.CreateL4HandlerRequest) (*andproto.CreateL4HandlerResponse, error) {

	funcName := "CreateL4Handler"

	response, err := config.DbMngClient.CreateL4Handler(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.CreateL4Handler()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	request.Hndlr.Id = response.Id

	hndlr := &l4handler.L4Handler{
		L4Handler: request.Hndlr,
	}

	if hndlr.Status == models.L4Handler_Listen {
		hndlr.StartListen()
	}

	l4handler.L4HandlersMutex.Lock()
	l4handler.L4Handlers[response.Id] = hndlr
	l4handler.L4HandlersMutex.Unlock()

	return response, nil
}

func (d *Proxy) UpdateL4Handler(ctx context.Context, request *andproto.UpdateL4HandlerRequest) (*emptypb.Empty, error) {

	funcName := "UpdateL4Handler"

	_, err := config.DbMngClient.UpdateL4Handler(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.UpdateL4Handler()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	l4handler.L4HandlersMutex.Lock()
	_, exists := l4handler.L4Handlers[request.Hndlr.Id]
	if exists {
		l4handler.L4Handlers[request.Hndlr.Id] = nil
	}
	l4handler.L4HandlersMutex.Unlock()

	hndlr := &l4handler.L4Handler{
		L4Handler: request.Hndlr,
	}

	if hndlr.Status == models.L4Handler_Listen {
		hndlr.StartListen()
	}

	l4handler.L4HandlersMutex.Lock()
	l4handler.L4Handlers[request.Hndlr.Id] = hndlr
	l4handler.L4HandlersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) DeleteL4Handler(ctx context.Context, request *andproto.DeleteL4HandlerRequest) (*emptypb.Empty, error) {

	funcName := "DeleteL4Handler"

	_, err := config.DbMngClient.DeleteL4Handler(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.DeleteL4Handler()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	l4handler.L4HandlersMutex.Lock()
	delete(l4handler.L4Handlers, request.Id)
	l4handler.L4HandlersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) ReadL4Handlers(ctx context.Context, request *andproto.ReadL4HandlersRequest) (*andproto.ReadL4HandlersResponse, error) {

	funcName := "ReadL4Handlers"

	resp, err := config.DbMngClient.ReadL4Handlers(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadL4Handlers()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}
	return resp, nil
}
