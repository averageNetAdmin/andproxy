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

var poolMutex = &sync.RWMutex{}
var pools = make(map[int64]*models.ServerPool)

func InitSrvPoolCashe() error {

	sql := `SELECT id, name, balancing_method FROM pool`
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		return err
	}
	defer rows.Close()

	subsql := `SELECT server_id, disabled_server FROM pool_servers WHERE pool_id = $1`

	for rows.Next() {
		item := &models.ServerPool{}

		err = rows.Scan(&item.Id, &item.Name, &item.BalancingMethod)
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
			srvId := int64(0)
			disabled := false

			err = rows.Scan(&srvId, &disabled)
			if err != nil {
				err = errors.WithMessage(err, "rows.Scan()")
				return err
			}

			if disabled {
				item.DisabledServers = append(item.DisabledServers, srvId)
			} else {
				item.Servers = append(item.Servers, srvId)
			}
		}

		pools[item.Id] = item
	}
	return nil
}

func (d *DbSrv) CreateServerPool(ctx context.Context, request *andproto.CreateServerPoolRequest) (*andproto.CreateServerPoolResponse, error) {

	funcName := "CreateServerPool()"

	response := &andproto.CreateServerPoolResponse{}

	sql := `INSERT INTO pool (name, balancing_method) VALUES ($1, $2) returning id`
	err := config.DB.QueryRow(sql, request.Pool.Name, request.Pool.BalancingMethod).Scan(&response.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", request.Pool.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO pool_servers (pool_id, server_id, disabled_server) VALUES ($1, $2, $3)`

	for _, srvId := range request.Pool.Servers {
		_, err := config.DB.Exec(subsql, response.Id, srvId, false)
		if err != nil {
			err = errors.WithMessage(err, "db.Exec()")
			config.Glogger.WithField("item", request.Pool.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	for _, srvId := range request.Pool.DisabledServers {
		_, err := config.DB.Exec(subsql, response.Id, srvId, true)
		if err != nil {
			err = errors.WithMessage(err, "db.Exec()")
			config.Glogger.WithField("item", request.Pool.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	request.Pool.Id = response.Id

	poolMutex.Lock()
	pools[response.Id] = request.Pool
	poolMutex.Unlock()

	return response, nil
}

func (d *DbSrv) UpdateServerPool(ctx context.Context, request *andproto.UpdateServerPoolRequest) (*emptypb.Empty, error) {

	funcName := "UpdateServerPool()"

	sql := `UPDATE INTO pool SET name = $1, balancing_method = $2 WHERE id = $3`
	_, err := config.DB.Exec(sql, request.Pool.Name, request.Pool.BalancingMethod, request.Pool.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Pool.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	tx, err := config.DB.Begin()
	if err != nil {
		err = errors.WithMessage(err, "db.Begin()")
		config.Glogger.WithField("item", request.Pool.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM pool_servers WHERE pool_id = $1`
	_, err = tx.Exec(sql, request.Pool.Id)
	if err != nil {
		err = errors.WithMessage(err, "tx.Exec()")
		config.Glogger.WithField("item", request.Pool.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO pool_servers (pool_id, server_id, disabled_server) VALUES ($1, $2, $3)`

	for _, srvId := range request.Pool.Servers {
		_, err := tx.Exec(subsql, request.Pool.Id, srvId, false)
		if err != nil {
			err = errors.WithMessage(err, "tx.Exec()")
			config.Glogger.WithField("item", request.Pool.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	for _, srvId := range request.Pool.DisabledServers {
		_, err := tx.Exec(subsql, request.Pool.Id, srvId, true)
		if err != nil {
			err = errors.WithMessage(err, "tx.Exec()")
			config.Glogger.WithField("item", request.Pool.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		err = errors.WithMessage(err, "tx.Commit()")
		config.Glogger.WithField("item", request.Pool.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	poolMutex.Lock()
	pools[request.Pool.Id] = request.Pool
	poolMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) DeleteServerPool(ctx context.Context, request *andproto.DeleteServerPoolRequest) (*emptypb.Empty, error) {

	funcName := "DeleteServerPool()"

	sql := `DELETE FROM pool WHERE id = $1`
	_, err := config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM pool_servers WHERE pool_id = $1`
	_, err = config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	poolMutex.Lock()
	delete(pools, request.Id)
	poolMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) ReadServerPools(ctx context.Context, request *andproto.ReadServerPoolsRequest) (*andproto.ReadServerPoolsResponse, error) {

	funcName := "ReadServerPools()"

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

	response := &andproto.ReadServerPoolsResponse{
		Pool: make([]*models.ServerPool, 0),
	}

	sql := fmt.Sprintf(`SELECT id, name, balancing_method FROM pool %s %s %s`, sortBy, limit, offset)
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}
	defer rows.Close()

	subsql := `SELECT server_id, disabled_server FROM pool_servers WHERE pool_id = $1`

	for rows.Next() {
		item := &models.ServerPool{}

		err = rows.Scan(&item.Id, &item.Name, &item.BalancingMethod)
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
			srvId := int64(0)
			disabled := false

			err = rows.Scan(&srvId, &disabled)
			if err != nil {
				err = errors.WithMessage(err, "rows.Scan()")
				config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
				return nil, err
			}

			if disabled {
				item.DisabledServers = append(item.DisabledServers, srvId)
			} else {
				item.Servers = append(item.Servers, srvId)
			}
		}

		response.Pool = append(response.Pool, item)
	}

	sql = `SELECT count(*) FROM pool`
	err = config.DB.QueryRow(sql).Scan(&response.RowsCount)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}

	return response, nil
}

func CheckServerPoolExists(id int64) bool {
	poolMutex.RLock()
	_, ok := pools[id]
	poolMutex.RUnlock()
	return ok
}
