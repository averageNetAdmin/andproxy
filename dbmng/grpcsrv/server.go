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

var serverMutex = &sync.RWMutex{}
var servers = make(map[int64]*models.Server)

func InitServerCashe() error {

	sql := `SELECT id, address, port, connect_timeout, weight, max_fails, break_time FROM server`
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		return err
	}
	defer rows.Close()

	for rows.Next() {
		item := &models.Server{}

		err = rows.Scan(&item.Id, &item.Address, &item.Port, &item.ConnectTimeout, &item.Weight, &item.MaxFails, &item.BreakTime)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			return err
		}
		servers[item.Id] = item
	}
	return nil
}

func (d *DbSrv) CreateServer(ctx context.Context, request *andproto.CreateServerRequest) (*andproto.CreateServerResponse, error) {

	funcName := "CreateServer()"

	response := &andproto.CreateServerResponse{}

	sql := `INSERT INTO server (address, port, connect_timeout, weight, max_fails, break_time) VALUES ($1, $2, $3, $4, $5, $6) returning id`
	err := config.DB.QueryRow(sql, request.Srv.Address, request.Srv.Port, request.Srv.ConnectTimeout,
		request.Srv.Weight, request.Srv.MaxFails, request.Srv.BreakTime).Scan(&response.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", fmt.Sprintf("%s:%d", request.Srv.Address, request.Srv.Port)).WithField("func", funcName).Error(err)
		return nil, err
	}

	request.Srv.Id = response.Id

	serverMutex.Lock()
	servers[response.Id] = request.Srv
	serverMutex.Unlock()

	return response, nil
}

func (d *DbSrv) UpdateServer(ctx context.Context, request *andproto.UpdateServerRequest) (*emptypb.Empty, error) {

	funcName := "UpdateServer()"

	sql := `UPDATE INTO server SET address = $1, port = $2, connect_timeout = $3, weight = $4, max_fails = $5, break_time= $6 WHERE id = $7`
	_, err := config.DB.Exec(sql, request.Srv.Address, request.Srv.Port, request.Srv.ConnectTimeout,
		request.Srv.Weight, request.Srv.MaxFails, request.Srv.BreakTime, request.Srv.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", fmt.Sprintf("%s:%d", request.Srv.Address, request.Srv.Port)).WithField("func", funcName).Error(err)
		return nil, err
	}

	serverMutex.Lock()
	servers[request.Srv.Id] = request.Srv
	serverMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) DeleteServer(ctx context.Context, request *andproto.DeleteServerRequest) (*emptypb.Empty, error) {

	funcName := "DeleteServer()"

	sql := `DELETE FROM server WHERE id = $1`
	_, err := config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	serverMutex.Lock()
	delete(servers, request.Id)
	serverMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) ReadServers(ctx context.Context, request *andproto.ReadServersRequest) (*andproto.ReadServersResponse, error) {

	funcName := "ReadServers()"

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

	response := &andproto.ReadServersResponse{
		Srv: make([]*models.Server, 0),
	}

	sql := fmt.Sprintf(`SELECT id, address, port, connect_timeout, weight, max_fails, break_time FROM server %s %s %s`, sortBy, limit, offset)
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &models.Server{}

		err = rows.Scan(&item.Id, &item.Address, &item.Port, &item.ConnectTimeout, &item.Weight, &item.MaxFails, &item.BreakTime)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}
		response.Srv = append(response.Srv, item)
	}

	sql = `SELECT count(*) FROM server`
	err = config.DB.QueryRow(sql).Scan(&response.RowsCount)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}

	return response, nil
}

func CheckServerExists(id int64) bool {
	serverMutex.RLock()
	_, ok := servers[id]
	serverMutex.RUnlock()
	return ok
}
