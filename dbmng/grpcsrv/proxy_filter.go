package grpcsrv

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/pkg/errors"

	"github.com/averageNetAdmin/andproxy/andproto"
	"github.com/averageNetAdmin/andproxy/andproto/models"
	"github.com/averageNetAdmin/andproxy/dbmng/config"
	"google.golang.org/protobuf/types/known/emptypb"
)

var filterMutex = &sync.RWMutex{}
var proxyFilter = make(map[int64]*models.ProxyFilter)

func InitFilterCashe() error {

	sql := `SELECT id, name, target_net, server_pool_id FROM proxy_filter`
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		return err
	}
	defer rows.Close()

	subsql := `SELECT acl_id, deny_acl FROM proxy_filter_acls WHERE filter_id = $1`

	for rows.Next() {
		item := &models.ProxyFilter{}

		err = rows.Scan(&item.Id, &item.Name, &item.TargetNet, &item.ServerPool)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			return err
		}

		rows, err := config.DB.Query(subsql, item.Id)
		if err != nil {
			err = errors.WithMessage(err, "db.Query()")
			return err
		}
		defer rows.Close()

		for rows.Next() {
			aclId := int64(0)
			deny := false

			err = rows.Scan(&aclId, &deny)
			if err != nil {
				err = errors.WithMessage(err, "rows.Scan()")
				return err
			}

			if deny {
				item.DenyAcls = append(item.DenyAcls, aclId)
			} else {
				item.AcceptAcls = append(item.AcceptAcls, aclId)
			}
		}

		proxyFilter[item.Id] = item
	}
	return nil
}

func (d *DbSrv) CreateProxyFilter(ctx context.Context, request *andproto.CreateProxyFilterRequest) (*andproto.CreateProxyFilterResponse, error) {

	funcName := "CreateProxyFilter()"

	response := &andproto.CreateProxyFilterResponse{}

	sql := `INSERT INTO proxy_filter (name, target_net, server_pool_id) VALUES ($1, $2, $3) returning id`
	err := config.DB.QueryRow(sql, request.Fltr.Name, request.Fltr.TargetNet, request.Fltr.ServerPool).Scan(&response.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", request.Fltr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO proxy_filter_acls (filter_id, acl_id, deny_acl) VALUES ($1, $2, $3)`

	for _, aclId := range request.Fltr.AcceptAcls {
		_, err := config.DB.Exec(subsql, response.Id, aclId, false)
		if err != nil {
			err = errors.WithMessage(err, "db.Exec()")
			config.Glogger.WithField("item", request.Fltr.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	for _, aclId := range request.Fltr.DenyAcls {
		_, err := config.DB.Exec(subsql, response.Id, aclId, true)
		if err != nil {
			err = errors.WithMessage(err, "db.Exec()")
			config.Glogger.WithField("item", request.Fltr.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	request.Fltr.Id = response.Id

	filterMutex.Lock()
	proxyFilter[response.Id] = request.Fltr
	filterMutex.Unlock()

	return response, nil
}

func (d *DbSrv) UpdateProxyFilter(ctx context.Context, request *andproto.UpdateProxyFilterRequest) (*emptypb.Empty, error) {

	funcName := "UpdateProxyFilter()"

	sql := `UPDATE INTO proxy_filter SET name = $1, target_net = $2, server_pool_id = $3 WHERE id = $4`
	_, err := config.DB.Exec(sql, request.Fltr.Name, request.Fltr.TargetNet, request.Fltr.ServerPool, request.Fltr.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Fltr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	tx, err := config.DB.Begin()
	if err != nil {
		err = errors.WithMessage(err, "db.Begin()")
		config.Glogger.WithField("item", request.Fltr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM proxy_filter_acls WHERE filter_id = $1`
	_, err = tx.Exec(sql, request.Fltr.Id)
	if err != nil {
		err = errors.WithMessage(err, "tx.Exec()")
		config.Glogger.WithField("item", request.Fltr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO proxy_filter_acls (filter_id, acl_id, deny_acl) VALUES ($1, $2, $3)`

	for _, aclId := range request.Fltr.AcceptAcls {
		_, err := tx.Exec(subsql, request.Fltr.Id, aclId, false)
		if err != nil {
			err = errors.WithMessage(err, "tx.Exec()")
			config.Glogger.WithField("item", request.Fltr.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	for _, aclId := range request.Fltr.DenyAcls {
		_, err := tx.Exec(subsql, request.Fltr.Id, aclId, true)
		if err != nil {
			err = errors.WithMessage(err, "tx.Exec()")
			config.Glogger.WithField("item", request.Fltr.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		err = errors.WithMessage(err, "tx.Commit()")
		config.Glogger.WithField("item", request.Fltr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	filterMutex.Lock()
	proxyFilter[request.Fltr.Id] = request.Fltr
	filterMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) DeleteProxyFilter(ctx context.Context, request *andproto.DeleteProxyFilterRequest) (*emptypb.Empty, error) {

	funcName := "DeleteProxyFilter()"

	sql := `DELETE FROM proxy_filter WHERE id = $1`
	_, err := config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM proxy_filter_acls WHERE filter_id = $1`
	_, err = config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	filterMutex.Lock()
	delete(proxyFilter, request.Id)
	filterMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) ReadProxyFilters(ctx context.Context, request *andproto.ReadProxyFiltersRequest) (*andproto.ReadProxyFiltersResponse, error) {

	funcName := "ReadProxyFilters()"

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

	response := &andproto.ReadProxyFiltersResponse{
		Fltr: make([]*models.ProxyFilter, 0),
	}

	sql := fmt.Sprintf(`SELECT id, name, target_net, server_pool_id FROM proxy_filter %s %s %s`, sortBy, limit, offset)
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}
	defer rows.Close()

	subsql := `SELECT acl_id, deny_acl FROM proxy_filter_acls WHERE filter_id = $1`

	for rows.Next() {
		item := &models.ProxyFilter{}

		err = rows.Scan(&item.Id, &item.Name, &item.TargetNet, &item.ServerPool)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}

		rows, err := config.DB.Query(subsql, item.Id)
		if err != nil {
			err = errors.WithMessage(err, "db.Query()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			aclId := int64(0)
			deny := false

			err = rows.Scan(&aclId, &deny)
			if err != nil {
				err = errors.WithMessage(err, "rows.Scan()")
				config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
				return nil, err
			}

			if deny {
				item.DenyAcls = append(item.DenyAcls, aclId)
			} else {
				item.AcceptAcls = append(item.AcceptAcls, aclId)
			}
		}

		response.Fltr = append(response.Fltr, item)
	}

	sql = `SELECT count(*) FROM proxy_filter`
	err = config.DB.QueryRow(sql).Scan(&response.RowsCount)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}

	return response, nil
}

func CheckProxyFilterExists(id int64) bool {
	filterMutex.RLock()
	_, ok := proxyFilter[id]
	filterMutex.RUnlock()
	return ok
}
