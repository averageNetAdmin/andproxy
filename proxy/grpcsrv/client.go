package grpcsrv

import (
	"context"
	"net"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/proxy/ac"
	"github.com/averageNetAdmin/andproxy/proxy/config"
	"github.com/averageNetAdmin/andproxy/proxy/tools"
	"google.golang.org/protobuf/types/known/emptypb"
)

func InitCashe() error {

	resp, err := config.DbMngClient.ReadAcls(context.Background(), &andproto.ReadAclsRequest{})
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadAcls()")
		return err
	}

	for _, item := range resp.Acl {

		nets := make([]*net.IPNet, 0)
		for _, n := range item.Nets {
			aces, err := tools.CreateRange(n)
			if err != nil {
				err = errors.WithMessage(err, "tools.CreateRange()")
				return err
			}
			for _, ace := range aces {
				// if string contains "/" it is network else address
				_, n, err := net.ParseCIDR(ace)
				if err != nil {
					err = errors.WithMessage(err, "net.ParseCIDR()")
					return err
				}

				nets = append(nets, n)
			}
		}
		ac.Acls[item.Id] = &ac.Acl{
			Acl:  item,
			Nets: nets,
		}
	}

	return nil
}

func (d *Proxy) CreateAcl(ctx context.Context, request *andproto.CreateAclRequest) (*andproto.CreateAclResponse, error) {

	funcName := "CreateAcl"

	response, err := config.DbMngClient.CreateAcl(ctx, request)
	if err != nil {
		return nil, err
	}

	nets := make([]*net.IPNet, 0)
	for _, n := range request.Acl.Nets {
		aces, err := tools.CreateRange(n)
		if err != nil {
			err = errors.WithMessage(err, "tools.CreateRange()")
			config.Glogger.WithField("item", request.Acl.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
		for _, ace := range aces {
			// if string contains "/" it is network else address
			_, n, err := net.ParseCIDR(ace)
			if err != nil {
				err = errors.WithMessage(err, "net.ParseCIDR()")
				config.Glogger.WithField("item", request.Acl.Name).WithField("func", funcName).Error(err)
				return nil, err
			}

			nets = append(nets, n)
		}
	}
	request.Acl.Id = response.Id

	ac.AclsMutex.Lock()
	ac.Acls[response.Id] = &ac.Acl{
		Acl:  request.Acl,
		Nets: nets,
	}
	ac.AclsMutex.Unlock()

	return response, nil
}

func (d *Proxy) UpdateAcl(ctx context.Context, request *andproto.UpdateAclRequest) (*emptypb.Empty, error) {

	funcName := "UpdateAcl"

	_, err := config.DbMngClient.UpdateAcl(ctx, request)
	if err != nil {
		return nil, err
	}

	nets := make([]*net.IPNet, 0)
	for _, n := range request.Acl.Nets {
		aces, err := tools.CreateRange(n)
		if err != nil {
			err = errors.WithMessage(err, "tools.CreateRange()")
			config.Glogger.WithField("item", request.Acl.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
		for _, ace := range aces {
			// if string contains "/" it is network else address
			_, n, err := net.ParseCIDR(ace)
			if err != nil {
				err = errors.WithMessage(err, "net.ParseCIDR()")
				config.Glogger.WithField("item", request.Acl.Name).WithField("func", funcName).Error(err)
				return nil, err
			}

			nets = append(nets, n)
		}
	}

	ac.AclsMutex.Lock()
	ac.Acls[request.Acl.Id] = &ac.Acl{
		Acl:  request.Acl,
		Nets: nets,
	}
	ac.AclsMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) DeleteAcl(ctx context.Context, request *andproto.DeleteAclRequest) (*emptypb.Empty, error) {

	funcName := "DeleteAcl"

	_, err := config.DbMngClient.DeleteAcl(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.DeleteAcl()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}

	ac.AclsMutex.Lock()
	delete(ac.Acls, request.Id)
	ac.AclsMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *Proxy) ReadAcls(ctx context.Context, request *andproto.ReadAclsRequest) (*andproto.ReadAclsResponse, error) {

	funcName := "ReadAcls"

	resp, err := config.DbMngClient.ReadAcls(ctx, request)
	if err != nil {
		err = errors.WithMessage(err, "DbMngClient.ReadHttpPaths()")
		config.Glogger.WithField("func", funcName).Error(err)
		return nil, err
	}
	return resp, nil
}
