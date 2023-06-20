package grpcsrv

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/andproto/models"
	"github.com/averageNetAdmin/andproxy/dbmng/config"
	"google.golang.org/protobuf/types/known/emptypb"
)

var aclMutex = &sync.RWMutex{}
var acls = make(map[int64]*models.Acl)

func InitCashe() error {

	sql := `SELECT id, name, nets  FROM acl`
	rows, err := config.DB.Query(sql)
	if err != nil {
		return errors.New("query server from db failed")
	}
	defer rows.Close()

	for rows.Next() {
		item := &models.Acl{}

		nets := ""
		err = rows.Scan(&item.Id, &item.Name, &nets)
		if err != nil {
			return errors.New("scan server rows from db failed")
		}
		netsArr := strings.Split(nets, ";")
		for _, n := range netsArr {
			_, net, err := net.ParseCIDR(n)
			if err != nil {
				err = errors.WithMessage(err, "net.ParseCIDR()")
				return err
			}
			item.Nets = append(item.Nets, net.String())
		}
		acls[item.Id] = item
	}
	return nil
}

func (d *DbSrv) CreateAcl(ctx context.Context, request *andproto.CreateAclRequest) (*andproto.CreateAclResponse, error) {

	funcName := "CreateAcl()"

	response := &andproto.CreateAclResponse{}

	strNets := ""
	for _, ace := range request.Acl.Nets {

		// if string contains "/" it is network else address
		_, _, err := net.ParseCIDR(ace)
		if err != nil {
			err = errors.WithMessage(err, "net.ParseCIDR()")
			config.Glogger.WithField("item", request.Acl.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
		strNets += ace + ";"
	}

	sql := `INSERT INTO acl (name, nets) VALUES ($1, $2) returning id`
	err := config.DB.QueryRow(sql, request.Acl.Name, strNets).Scan(&response.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", request.Acl.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	request.Acl.Id = response.Id

	aclMutex.Lock()
	acls[response.Id] = request.Acl
	aclMutex.Unlock()

	return response, nil
}

func (d *DbSrv) UpdateAcl(ctx context.Context, request *andproto.UpdateAclRequest) (*emptypb.Empty, error) {

	funcName := "UpdateAcl()"

	strNets := ""
	for _, ace := range request.Acl.Nets {

		// if string contains "/" it is network else address
		_, _, err := net.ParseCIDR(ace)
		if err != nil {
			err = errors.WithMessage(err, "net.ParseCIDR()")
			config.Glogger.WithField("item", request.Acl.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
		strNets += ace + ";"
	}
	sql := `UPDATE INTO acl SET name = $1, nets = $2 WHERE id = $3`
	_, err := config.DB.Exec(sql, request.Acl.Name, strNets, request.Acl.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Acl.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	aclMutex.Lock()
	acls[request.Acl.Id] = request.Acl
	aclMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) DeleteAcl(ctx context.Context, request *andproto.DeleteAclRequest) (*emptypb.Empty, error) {

	funcName := "DeleteAcl()"

	sql := `DELETE FROM acl WHERE id = $1`
	_, err := config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	aclMutex.Lock()
	delete(acls, request.Id)
	aclMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) ReadAcls(ctx context.Context, request *andproto.ReadAclsRequest) (*andproto.ReadAclsResponse, error) {

	funcName := "ReadAcls()"

	if request.SortBy == "" {
		request.SortBy = "id"
	}
	sortBy := fmt.Sprintf("ORDER BY %s ", request.SortBy)

	limit := ""
	offset := ""
	if request.Limit > 0 {
		limit = fmt.Sprintf("LIMIT %s ", strconv.Itoa(int(request.Limit)))
		offset = fmt.Sprintf("OFFSET %s ", strconv.Itoa(int(request.Offset)))
	}

	response := &andproto.ReadAclsResponse{
		Acl: make([]*models.Acl, 0),
	}

	sql := `SELECT id, name, nets FROM acl` + sortBy + limit + offset
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &models.Acl{}

		nets := ""
		err = rows.Scan(&item.Id, &item.Name, &nets)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}
		netsArr := strings.Split(nets, ";")
		for _, n := range netsArr {
			_, net, err := net.ParseCIDR(n)
			if err != nil {
				err = errors.WithMessage(err, "net.ParseCIDR()")
				config.Glogger.WithField("item", item.Name).WithField("func", funcName).Error(err)
				return nil, err
			}
			item.Nets = append(item.Nets, net.String())
		}
		response.Acl = append(response.Acl, item)
	}

	sql = `SELECT count(*) FROM acl`
	err = config.DB.QueryRow(sql).Scan(&response.RowsCount)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}

	return response, nil
}

func CheckAclExists(id int64) bool {
	aclMutex.RLock()
	_, ok := acls[id]
	aclMutex.RUnlock()
	return ok
}
