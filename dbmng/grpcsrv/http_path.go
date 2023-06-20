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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

var httpPathMutex = &sync.RWMutex{}
var httpPath = make(map[int64]*models.HttpPath)

func InitHttpPathCashe() error {

	sql := `SELECT id, path, enable_caching, request_timeout, max_connections, set_input, set_output, statuc_response, type, disabled FROM http_path`
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		return err
	}
	defer rows.Close()

	for rows.Next() {
		item := &models.HttpPath{}
		setOut := []byte("")
		setIn := []byte("")
		respSt := []byte("")
		disabled := false

		err = rows.Scan(&item.Id, &item.Path, &item.EnableCaching, &item.RequestTimeout, &item.MaxConnections, &setIn, &setOut,
			&respSt, &item.Type, &disabled)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			return err
		}

		err := proto.Unmarshal(setIn, item.SetInput)
		if err != nil {
			err = errors.WithMessage(err, "proto.Unmarshal()")
			return err
		}

		err = proto.Unmarshal(setOut, item.SetOutput)
		if err != nil {
			err = errors.WithMessage(err, "proto.Unmarshal()")
			return err
		}

		err = proto.Unmarshal(respSt, item.StaticResponse)
		if err != nil {
			err = errors.WithMessage(err, "proto.Unmarshal()")
			return err
		}

		if disabled {
			item.Status = models.HttpPath_Disabled
		}

		httpPath[item.Id] = item
	}
	return nil
}

func (d *DbSrv) CreateHttpPath(ctx context.Context, request *andproto.CreateHttpPathRequest) (*andproto.CreateHttpPathResponse, error) {

	funcName := "CreateHttpPath()"

	response := &andproto.CreateHttpPathResponse{}
	disabled := request.Path.Status == models.HttpPath_Disabled

	setIn, err := proto.Marshal(request.Path.SetInput)
	if err != nil {
		err = errors.WithMessage(err, "proto.Marshal()")
		config.Glogger.WithField("item", request.Path.Path).WithField("func", funcName).Error(err)
		return nil, err
	}

	setOut, err := proto.Marshal(request.Path.SetOutput)
	if err != nil {
		err = errors.WithMessage(err, "proto.Marshal()")
		config.Glogger.WithField("item", request.Path.Path).WithField("func", funcName).Error(err)
		return nil, err
	}

	respSt, err := proto.Marshal(request.Path.StaticResponse)
	if err != nil {
		err = errors.WithMessage(err, "proto.Marshal()")
		config.Glogger.WithField("item", request.Path.Path).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql := `INSERT INTO http_path (path, enable_caching, request_timeout, max_connections, set_input, set_output, statuc_response, type, disabled) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`
	err = config.DB.QueryRow(sql, request.Path.Path, request.Path.EnableCaching, request.Path.RequestTimeout, request.Path.MaxConnections,
		setIn, setOut, respSt, request.Path.Type, disabled).Scan(&response.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", request.Path.Path).WithField("func", funcName).Error(err)
		return nil, err
	}

	request.Path.Id = response.Id

	httpPathMutex.Lock()
	httpPath[response.Id] = request.Path
	httpPathMutex.Unlock()

	return response, nil
}

func (d *DbSrv) UpdateHttpPath(ctx context.Context, request *andproto.UpdateHttpPathRequest) (*emptypb.Empty, error) {

	funcName := "UpdateHttpPath()"

	disabled := request.Path.Status == models.HttpPath_Disabled

	setIn, err := proto.Marshal(request.Path.SetInput)
	if err != nil {
		err = errors.WithMessage(err, "proto.Marshal()")
		config.Glogger.WithField("item", request.Path.Path).WithField("func", funcName).Error(err)
		return nil, err
	}

	setOut, err := proto.Marshal(request.Path.SetOutput)
	if err != nil {
		err = errors.WithMessage(err, "proto.Marshal()")
		config.Glogger.WithField("item", request.Path.Path).WithField("func", funcName).Error(err)
		return nil, err
	}

	respSt, err := proto.Marshal(request.Path.StaticResponse)
	if err != nil {
		err = errors.WithMessage(err, "proto.Marshal()")
		config.Glogger.WithField("item", request.Path.Path).WithField("func", funcName).Error(err)
		return nil, err
	}

	sql := `UPDATE INTO http_path SET path = $1, enable_caching = $2, request_timeout = $3, max_connections = $4, set_input = $5, 
	set_output = $6, statuc_response = $7, type = $8, disabled = $9 WHERE id = $10`
	_, err = config.DB.Exec(sql, request.Path.Path, request.Path.EnableCaching, request.Path.RequestTimeout, request.Path.MaxConnections,
		setIn, setOut, respSt, request.Path.Type, disabled, request.Path.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Path.Path).WithField("func", funcName).Error(err)
		return nil, err
	}

	httpPathMutex.Lock()
	httpPath[request.Path.Id] = request.Path
	httpPathMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) DeleteHttpPath(ctx context.Context, request *andproto.DeleteHttpPathRequest) (*emptypb.Empty, error) {

	funcName := "DeleteHttpPath()"

	sql := `DELETE FROM http_path WHERE id = $1`
	_, err := config.DB.Exec(sql, request.Id)
	if err != nil {
		err = errors.WithMessage(err, "db.Exec()")
		config.Glogger.WithField("item", request.Id).WithField("func", funcName).Error(err)
		return nil, err
	}

	httpPathMutex.Lock()
	delete(httpPath, request.Id)
	httpPathMutex.Unlock()

	return &emptypb.Empty{}, nil
}

func (d *DbSrv) ReadHttpPaths(ctx context.Context, request *andproto.ReadHttpPathsRequest) (*andproto.ReadHttpPathsResponse, error) {

	funcName := "ReadHttpPaths()"

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

	response := &andproto.ReadHttpPathsResponse{
		Path: make([]*models.HttpPath, 0),
	}

	sql := fmt.Sprintf(`SELECT id, path, enable_caching, request_timeout, max_connections, set_input, 
		set_output, statuc_response, type, disabled FROM http_path %s %s %s`, sortBy, limit, offset)
	rows, err := config.DB.Query(sql)
	if err != nil {
		err = errors.WithMessage(err, "db.Query()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &models.HttpPath{}

		disabled := false
		setOut := []byte("")
		setIn := []byte("")
		respSt := []byte("")

		err = rows.Scan(&item.Id, &item.Path, &item.EnableCaching, &item.RequestTimeout, &item.MaxConnections, &setIn, &setOut,
			&respSt, &item.Type, &disabled)
		if err != nil {
			err = errors.WithMessage(err, "rows.Scan()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}

		err := proto.Unmarshal(setIn, item.SetInput)
		if err != nil {
			err = errors.WithMessage(err, "proto.Unmarshal()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}

		err = proto.Unmarshal(setOut, item.SetOutput)
		if err != nil {
			err = errors.WithMessage(err, "proto.Unmarshal()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}

		err = proto.Unmarshal(respSt, item.StaticResponse)
		if err != nil {
			err = errors.WithMessage(err, "proto.Unmarshal()")
			config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
			return nil, err
		}

		if disabled {
			item.Status = models.HttpPath_Disabled
		}

		response.Path = append(response.Path, item)
	}

	sql = `SELECT count(*) FROM http_path`
	err = config.DB.QueryRow(sql).Scan(&response.RowsCount)
	if err != nil {
		err = errors.WithMessage(err, "db.QueryRow()")
		config.Glogger.WithField("item", "").WithField("func", funcName).Error(err)
		return nil, err
	}

	return response, nil
}

func CheckHttpPathExists(id int64) bool {
	httpPathMutex.RLock()
	_, ok := httpPath[id]
	httpPathMutex.RUnlock()
	return ok
}
