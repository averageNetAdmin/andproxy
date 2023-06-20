package config

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

var DB *sql.DB

func OpenDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "./andproxy.db")
	if err != nil {
		err = errors.WithMessage(err, "sql.Open()")
		return err
	}
	ctx := context.Background()
	dt, err := os.ReadFile("./db.sql")
	if err != nil {
		err = errors.WithMessage(err, "os.ReadFile()")
		return err
	}

	_, err = DB.ExecContext(ctx, string(dt))
	if err != nil {
		err = errors.WithMessage(err, "DB.ExecContext()")
		return err
	}
	return nil
}
