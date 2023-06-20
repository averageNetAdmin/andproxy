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

var httpHostMutex = &sync.RWMutex{}
var httpHost = make(map[int64]*models.HttpHost)

func InitHttpHostCashe() error {

	sql := `SELECT id, address, cert_path, cert_key_path, max_connections, disabled FROM http_host`
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		return err
	}
	defer rows.Close()

	subsql := `SELECT path_id FROM http_host_http_paths WHERE host_id = $1`

	for rows.Next() {
		item := &models.HttpHost{}
		disabled := false

		err = rows.Scan(&item.Id, &item.Address, &item.CertPath, &item.CertKeyPath, &item.MaxConnections, &disabled)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			return err
		}

		if disabled {
			item.Status = models.HttpHost_Disabled
		}

		rows, err := config.DB.Query(subsql, item.Id)
		if err != nil {
			err = errors.WithMessage(err, "db.Query()")
			return err
		}
		defer rows.Close()

		for rows.Next() {
			pathId := int64(0)

			err = rows.Scan(&pathId)
			if err != nil {
				err = errors.WithMessage(err, "rows.Scan()")
				return err
			}

			item.Paths = append(item.Paths, pathId)
		}

		httpHost[item.Id] = item
	}
	return nil
}

func (d *DbSrv) CreateHttpHost(ctx context.Context, request *andproto.CreateHttpHostRequest) (*andproto.CreateHttpHostResponse, error) {

	funcName := "CreateHttpHost()"

	response := &andproto.CreateHttpHostResponse{}

	disabled := request.Host.Status == models.HttpHost_Disabled

	sql := `INSERT INTO http_host (address, cert_path, cert_key_path, max_connections, disabled) 
		VALUES ($1, $2, $3, $4, $5) returning id`
	err := config.DB.QueryRow(sql, request.Host.Address, request.Host.CertPath, request.Host.CertKeyPath,
		request.Host.MaxConnections, disabled).Scan(&response.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", request.Host.Address).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO http_host_http_paths (host_id, path_id) VALUES ($1, $2)`

	for _, pathId := range request.Host.Paths {
		_, err := config.DB.Exec(subsql, response.Id, pathId)
		if err != nil {
			err = errors.WithMessage(err, "db.Exec()")
			config.Glogger.WithField("item", request.Host.Address).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	request.Host.Id = response.Id

	httpHostMutex.Lock()
	httpHost[response.Id] = request.Host
	httpHostMutex.Unlock()

	return response, nil
}

func (d *DbSrv) UpdateHttpHost(ctx context.Context, request *andproto.UpdateHttpHostRequest) (*emptypb.Empty, error) {

	funcName := "UpdateHttpHost()"

	disabled := request.Host.Status == models.HttpHost_Disabled

	sql := `UPDATE INTO http_host SET address = $1, cert_path = $2, cert_key_path = $3, max_connections = $4, disabled = $5 WHERE id = $6`
	_, err := config.DB.Exec(sql, request.Host.Address, request.Host.CertPath, request.Host.CertKeyPath,
		request.Host.MaxConnections, disabled, request.Host.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Host.Address).WithField("func", funcName).Error(err)
		return nil, err
	}

	tx, err := config.DB.Begin()
	if err != nil {
		err = errors.WithMessage(err, "db.Begin()")
		config.Glogger.WithField("item", request.Host.Address).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM http_host_http_paths WHERE host_id = $1`
	_, err = tx.Exec(sql, request.Host.Id)
	if err != nil {
		err = errors.WithMessage(err, "tx.Exec()")
		config.Glogger.WithField("item", request.Host.Address).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO http_host_http_paths (host_id, path_id) VALUES ($1, $2)`

	for _, pathId := range request.Host.Paths {
		_, err := tx.Exec(subsql, request.Host.Id, pathId)
		if err != nil {
			err = errors.WithMessage(err, "tx.Exec()")
			config.Glogger.WithField("item", request.Host.Address).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		err = errors.WithMessage(err, "tx.Commit()")
		config.Glogger.WithField("item", request.Host.Address).WithField("func", funcName).Error(err)
		return nil, err
	}

	httpHostMutex.Lock()
	httpHost[request.Host.Id] = request.Host
	httpHostMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) DeleteHttpHost(ctx context.Context, request *andproto.DeleteHttpHostRequest) (*emptypb.Empty, error) {

	funcName := "DeleteHttpHost()"

	sql := `DELETE FROM http_host WHERE id = $1`
	_, err := config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM http_host_http_paths WHERE host_id = $1`
	_, err = config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	httpHostMutex.Lock()
	delete(httpHost, request.Id)
	httpHostMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) ReadHttpHosts(ctx context.Context, request *andproto.ReadHttpHostsRequest) (*andproto.ReadHttpHostsResponse, error) {

	funcName := "ReadHttpHosts()"

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

	response := &andproto.ReadHttpHostsResponse{
		Host: make([]*models.HttpHost, 0),
	}

	sql := fmt.Sprintf(`SELECT id, address, cert_path, cert_key_path, max_connections, disabled FROM http_host %s %s %s`, sortBy, limit, offset)
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}
	defer rows.Close()

	subsql := `SELECT path_id FROM http_host_http_paths WHERE host_id = $1`

	for rows.Next() {
		item := &models.HttpHost{}

		disabled := false

		err = rows.Scan(&item.Id, &item.Address, &item.CertPath, &item.CertKeyPath, &item.MaxConnections, &disabled)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}

		if disabled {
			item.Status = models.HttpHost_Disabled
		}

		rows, err := config.DB.Query(subsql, item.Id)
		if err != nil {
			err = errors.WithMessage(err, "db.Query()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			pathId := int64(0)

			err = rows.Scan(&pathId)
			if err != nil {
				err = errors.WithMessage(err, "rows.Scan()")
				config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
				return nil, err
			}

			item.Paths = append(item.Paths, pathId)
		}

		response.Host = append(response.Host, item)
	}

	sql = `SELECT count(*) FROM http_host`
	err = config.DB.QueryRow(sql).Scan(&response.RowsCount)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}

	return response, nil
}

func CheckHttpHostExists(id int64) bool {
	httpHostMutex.RLock()
	_, ok := httpHost[id]
	httpHostMutex.RUnlock()
	return ok
}
