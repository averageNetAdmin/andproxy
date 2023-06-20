package grpcsrv

import (
	"context"
	"crypto/tls"
	"regexp"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/handler/httph"
	"google.golang.org/protobuf/types/known/emptypb"
)

func InitHttpHostCashe() error {

	resp, err := config.DbMngClient.ReadHttpHosts(context.Background(), &andproto.ReadHttpHostsRequest{})
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadHttpHosts()")
		return err
	}

	for _, item := range resp.Host {
		certificate := tls.Certificate{}
		if item.CertPath != "" {
			certificate, err = tls.LoadX509KeyPair(item.CertPath, item.CertKeyPath)
			if err != nil {
				return err
			}
		}

		reg, err := regexp.Compile(item.Address)
		if err != nil {
			err = errors.WithMessage(err, "regexp.Compile()")
			return err
		}

		httph.HttpHosts[item.Id] = &httph.HttpHost{
			HttpHost:    item,
			Certificate: &certificate,
			HostReg:     reg,
		}
	}

	return nil
}

func (d *Proxy) CreateHttpHost(ctx context.Context, request *andproto.CreateHttpHostRequest) (*andproto.CreateHttpHostResponse, error) {

	funcName := "CreateHttpHost"

	var err error

	certificate := tls.Certificate{}
	if request.Host.CertPath != "" {
		certificate, err = tls.LoadX509KeyPair(request.Host.CertPath, request.Host.CertKeyPath)
		if err != nil {
			err = errors.WithMessage(err, "tls.LoadX509KeyPair()")
			config.Glogger.WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	reg, err := regexp.Compile(request.Host.Address)
	if err != nil {
		err = errors.WithMessage(err, "regexp.Compile()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	response, err := config.DbMngClient.CreateHttpHost(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.CreateHttpHost()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	request.Host.Id = response.Id

	httph.HttpHostsMutex.Lock()
	httph.HttpHosts[response.Id] = &httph.HttpHost{
		HttpHost:    request.Host,
		HostReg:     reg,
		Certificate: &certificate,
	}
	httph.HttpHostsMutex.Unlock()

	return response, nil
}

func (d *Proxy) UpdateHttpHost(ctx context.Context, request *andproto.UpdateHttpHostRequest) (*emptypb.Empty, error) {

	funcName := "UpdateHttpHost"

	var err error

	certificate := tls.Certificate{}
	if request.Host.CertPath != "" {
		certificate, err = tls.LoadX509KeyPair(request.Host.CertPath, request.Host.CertKeyPath)
		if err != nil {
			err = errors.WithMessage(err, "tls.LoadX509KeyPair()")
			config.Glogger.WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	reg, err := regexp.Compile(request.Host.Address)
	if err != nil {
		err = errors.WithMessage(err, "regexp.Compile()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	_, err = config.DbMngClient.UpdateHttpHost(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.UpdateHttpHost()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	httph.HttpHostsMutex.Lock()
	httph.HttpHosts[request.Host.Id] = &httph.HttpHost{
		HttpHost:    request.Host,
		HostReg:     reg,
		Certificate: &certificate,
	}
	httph.HttpHostsMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) DeleteHttpHost(ctx context.Context, request *andproto.DeleteHttpHostRequest) (*emptypb.Empty, error) {

	funcName := "DeleteHttpHost"

	_, err := config.DbMngClient.DeleteHttpHost(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.DeleteHttpHost()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	httph.HttpHostsMutex.Lock()
	delete(httph.HttpHosts, request.Id)
	httph.HttpHostsMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) ReadHttpHosts(ctx context.Context, request *andproto.ReadHttpHostsRequest) (*andproto.ReadHttpHostsResponse, error) {

	funcName := "ReadHttpHosts"

	resp, err := config.DbMngClient.ReadHttpHosts(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadHttpHosts()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}
	return resp, nil
}
