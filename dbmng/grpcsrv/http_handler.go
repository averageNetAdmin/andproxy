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

var httpHandlerMutex = &sync.RWMutex{}
var httpHandler = make(map[int64]*models.HttpHandler)

func InitHttpHandlerCashe() error {

	sql := `SELECT id, name, port, secure, max_connections, disabled FROM http_handler`
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		return err
	}
	defer rows.Close()

	subsql := `SELECT host_id FROM http_handler_http_hosts WHERE handler_id = $1`

	for rows.Next() {
		item := &models.HttpHandler{}
		disabled := false

		err = rows.Scan(&item.Id, &item.Name, &item.Port, &item.Secure, &item.MaxConnections, &disabled)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			return err
		}

		if disabled {
			item.Status = models.HttpHandler_Stopped
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

			item.Hosts = append(item.Hosts, pathId)
		}

		httpHandler[item.Id] = item
	}
	return nil
}

func (d *DbSrv) CreateHttpHandler(ctx context.Context, request *andproto.CreateHttpHandlerRequest) (*andproto.CreateHttpHandlerResponse, error) {

	funcName := "CreateHttpHandler()"

	response := &andproto.CreateHttpHandlerResponse{}

	disabled := request.Hndlr.Status == models.HttpHandler_Stopped

	sql := `INSERT INTO http_handler (name, port, secure, max_connections, disabled) 
		VALUES ($1, $2, $3, $4, $5) returning id`
	err := config.DB.QueryRow(sql, request.Hndlr.Name, request.Hndlr.Port, request.Hndlr.Secure,
		request.Hndlr.MaxConnections, disabled).Scan(&response.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO http_handler_http_hosts (handler_id, host_id) VALUES ($1, $2)`

	for _, pathId := range request.Hndlr.Hosts {
		_, err := config.DB.Exec(subsql, response.Id, pathId)
		if err != nil {
			err = errors.WithMessage(err, "db.Exec()")
			config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	request.Hndlr.Id = response.Id

	httpHandlerMutex.Lock()
	httpHandler[response.Id] = request.Hndlr
	httpHandlerMutex.Unlock()

	return response, nil
}

func (d *DbSrv) UpdateHttpHandler(ctx context.Context, request *andproto.UpdateHttpHandlerRequest) (*emptypb.Empty, error) {

	funcName := "UpdateHttpHandler()"

	disabled := request.Hndlr.Status == models.HttpHandler_Stopped

	sql := `UPDATE INTO http_handler SET name = $1, port = $2, secure = $3, max_connections = $4, disabled = $5 WHERE id = $6`
	_, err := config.DB.Exec(sql, request.Hndlr.Name, request.Hndlr.Port, request.Hndlr.Secure,
		request.Hndlr.MaxConnections, disabled, request.Hndlr.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	tx, err := config.DB.Begin()
	if err != nil {
		err = errors.WithMessage(err, "tx.Begin()")
		config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM http_handler_http_hosts WHERE handler_id = $1`
	_, err = tx.Exec(sql, request.Hndlr.Id)
	if err != nil {
		err = errors.WithMessage(err, "tx.Exec()")
		config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO http_handler_http_hosts (handler_id, host_id) VALUES ($1, $2)`

	for _, pathId := range request.Hndlr.Hosts {
		_, err := tx.Exec(subsql, request.Hndlr.Id, pathId)
		if err != nil {
			err = errors.WithMessage(err, "tx.Exec()")
			config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	err = tx.Commit()
	if err != nil {
		err = errors.WithMessage(err, "tx.Commit()")
		config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	httpHandlerMutex.Lock()
	httpHandler[request.Hndlr.Id] = request.Hndlr
	httpHandlerMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) DeleteHttpHandler(ctx context.Context, request *andproto.DeleteHttpHandlerRequest) (*emptypb.Empty, error) {

	funcName := "DeleteHttpHandler()"

	sql := `DELETE FROM http_handler WHERE id = $1`
	_, err := config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM http_handler_http_hosts WHERE handler_id = $1`
	_, err = config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	httpHandlerMutex.Lock()
	delete(httpHandler, request.Id)
	httpHandlerMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) ReadHttpHandlers(ctx context.Context, request *andproto.ReadHttpHandlersRequest) (*andproto.ReadHttpHandlersResponse, error) {

	funcName := "ReadHttpHandlers()"

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

	response := &andproto.ReadHttpHandlersResponse{
		Hndlr: make([]*models.HttpHandler, 0),
	}

	sql := fmt.Sprintf(`SELECT id, name, port, secure, max_connections, disabled FROM http_handler %s %s %s`, sortBy, limit, offset)
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}
	defer rows.Close()

	subsql := `SELECT host_id FROM http_handler_http_hosts WHERE handler_id = $1`

	for rows.Next() {
		item := &models.HttpHandler{}

		disabled := false

		err = rows.Scan(&item.Id, &item.Name, &item.Port, &item.Secure, &item.MaxConnections, &disabled)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}

		if disabled {
			item.Status = models.HttpHandler_Stopped
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

			item.Hosts = append(item.Hosts, pathId)
		}

		response.Hndlr = append(response.Hndlr, item)
	}

	sql = `SELECT count(*) FROM http_handler`
	err = config.DB.QueryRow(sql).Scan(&response.RowsCount)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}

	return response, nil
}
