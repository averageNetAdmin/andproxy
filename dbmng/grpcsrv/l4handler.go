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

var l4handlersMutex = &sync.RWMutex{}
var l4handlers = make(map[int64]*models.L4Handler)

func InitL4HandlerCashe() error {

	sql := `SELECT id, name, protocol, port, deadline, write_deadLine, read_deadLine, max_connections, disabled FROM l4handler`
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		return err
	}
	defer rows.Close()

	subsql := `SELECT proxy_filter_id, disabled_filter FROM l4handler_proxy_filters WHERE hanlder_id = $1`

	for rows.Next() {
		item := &models.L4Handler{}
		disabled := false

		err = rows.Scan(&item.Id, &item.Name, &item.Protocol, &item.Port, &item.Deadline, &item.WriteDeadline,
			&item.ReadDeadline, &item.MaxConnections, &disabled)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			return err
		}

		if disabled {
			item.Status = models.L4Handler_Stopped
		}

		rows, err := config.DB.Query(subsql, item.Id)
		if err != nil {
			err = errors.WithMessage(err, "db.Query()")
			return err
		}
		defer rows.Close()

		for rows.Next() {
			filterId := int64(0)
			disable := false

			err = rows.Scan(&filterId, &disable)
			if err != nil {
				err = errors.WithMessage(err, "rows.Scan()")
				return err
			}

			if disable {
				item.DisabledProxyFilters = append(item.DisabledProxyFilters, filterId)
			} else {
				item.ProxyFilters = append(item.ProxyFilters, filterId)
			}
		}

		l4handlers[item.Id] = item
	}
	return nil
}

func (d *DbSrv) CreateL4Handler(ctx context.Context, request *andproto.CreateL4HandlerRequest) (*andproto.CreateL4HandlerResponse, error) {

	funcName := "CreateL4Handler()"

	response := &andproto.CreateL4HandlerResponse{}

	disabled := request.Hndlr.Status == models.L4Handler_Stopped

	sql := `INSERT INTO l4handler (name, protocol, port, deadline, write_deadLine, read_deadLine, max_connections, disabled) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) returning id`
	err := config.DB.QueryRow(sql, request.Hndlr.Name, request.Hndlr.Protocol, request.Hndlr.Port, request.Hndlr.Deadline,
		request.Hndlr.WriteDeadline, request.Hndlr.ReadDeadline, request.Hndlr.MaxConnections, disabled).Scan(&response.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO l4handler_proxy_filters (hanlder_id, proxy_filter_id, disabled_filter) VALUES ($1, $2, $3)`

	for _, filterId := range request.Hndlr.ProxyFilters {
		_, err := config.DB.Exec(subsql, response.Id, filterId, false)
		if err != nil {
			err = errors.WithMessage(err, "db.Exec()")
			config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	for _, filterId := range request.Hndlr.DisabledProxyFilters {
		_, err := config.DB.Exec(subsql, response.Id, filterId, true)
		if err != nil {
			err = errors.WithMessage(err, "db.Exec()")
			config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	request.Hndlr.Id = response.Id

	l4handlersMutex.Lock()
	l4handlers[response.Id] = request.Hndlr
	l4handlersMutex.Unlock()

	return response, nil
}

func (d *DbSrv) UpdateL4Handler(ctx context.Context, request *andproto.UpdateL4HandlerRequest) (*emptypb.Empty, error) {

	funcName := "UpdateL4Handler()"

	disabled := request.Hndlr.Status == models.L4Handler_Stopped

	sql := `UPDATE INTO l4handler SET name = $1, protocol = $2, port = $3, deadline = $4, write_deadLine = $5, 
		read_deadLine = $6, max_connections = $7, disabled = $8 WHERE id = $9`
	_, err := config.DB.Exec(sql, request.Hndlr.Name, request.Hndlr.Protocol, request.Hndlr.Port, request.Hndlr.Deadline,
		request.Hndlr.WriteDeadline, request.Hndlr.ReadDeadline, request.Hndlr.MaxConnections, disabled, request.Hndlr.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	tx, err := config.DB.Begin()
	if err != nil {
		err = errors.WithMessage(err, "db.Begin()")
		config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM l4handler_proxy_filters WHERE hanlder_id = $1`
	_, err = tx.Exec(sql, request.Hndlr.Id)
	if err != nil {
		err = errors.WithMessage(err, "tx.Exec()")
		config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
		return nil, err
	}

	subsql := `INSERT INTO l4handler_proxy_filters (hanlder_id, proxy_filter_id, disabled_filter) VALUES ($1, $2, $3)`

	for _, filterId := range request.Hndlr.ProxyFilters {
		_, err := tx.Exec(subsql, request.Hndlr.Id, filterId, false)
		if err != nil {
			err = errors.WithMessage(err, "tx.Exec()")
			config.Glogger.WithField("item", request.Hndlr.Name).WithField("func", funcName).Error(err)
			return nil, err
		}
	}

	for _, filterId := range request.Hndlr.DisabledProxyFilters {
		_, err := tx.Exec(subsql, request.Hndlr.Id, filterId, true)
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

	l4handlersMutex.Lock()
	l4handlers[request.Hndlr.Id] = request.Hndlr
	l4handlersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) DeleteL4Handler(ctx context.Context, request *andproto.DeleteL4HandlerRequest) (*emptypb.Empty, error) {

	funcName := "DeleteL4Handler()"

	sql := `DELETE FROM l4handler WHERE id = $1`
	_, err := config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql = `DELETE FROM l4handler_proxy_filters WHERE hanlder_id = $1`
	_, err = config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	l4handlersMutex.Lock()
	delete(l4handlers, request.Id)
	l4handlersMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) ReadL4Handlers(ctx context.Context, request *andproto.ReadL4HandlersRequest) (*andproto.ReadL4HandlersResponse, error) {

	funcName := "ReadL4Handlers()"

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

	response := &andproto.ReadL4HandlersResponse{
		Hndlr: make([]*models.L4Handler, 0),
	}

	sql := fmt.Sprintf(`SELECT id, name, protocol, port, deadline, write_deadLine, read_deadLine, max_connections, disabled FROM l4handler %s %s %s`, sortBy, limit, offset)
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}
	defer rows.Close()

	subsql := `SELECT proxy_filter_id, disabled_filter FROM l4handler_proxy_filters WHERE hanlder_id = $1`

	for rows.Next() {
		item := &models.L4Handler{}

		disabled := false

		err = rows.Scan(&item.Id, &item.Name, &item.Protocol, &item.Port, &item.Deadline, &item.WriteDeadline,
			&item.ReadDeadline, &item.MaxConnections, &disabled)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}

		if disabled {
			item.Status = models.L4Handler_Stopped
		}

		rows, err := config.DB.Query(subsql, item.Id)
		if err != nil {
			err = errors.WithMessage(err, "db.Query()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			filterId := int64(0)
			disable := false

			err = rows.Scan(&filterId, &disable)
			if err != nil {
				err = errors.WithMessage(err, "rows.Scan()")
				config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
				return nil, err
			}

			if disable {
				item.DisabledProxyFilters = append(item.DisabledProxyFilters, filterId)
			} else {
				item.ProxyFilters = append(item.ProxyFilters, filterId)
			}
		}

		response.Hndlr = append(response.Hndlr, item)
	}

	sql = `SELECT count(*) FROM l4handler`
	err = config.DB.QueryRow(sql).Scan(&response.RowsCount)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}

	return response, nil
}

func CheckL4HandlerExists(id int64) bool {
	l4handlersMutex.RLock()
	_, ok := l4handlers[id]
	l4handlersMutex.RUnlock()
	return ok
}
