package config

// import (
// 	"context"
// 	"database/sql"
// 	"os"

// 	_ "github.com/mattn/go-sqlite3"
// )

// var DB *sql.DB

// func OpenDB() error {
// 	var err error
// 	DB, err = sql.Open("sqlite3", "./andproxy.db")
// 	if err != nil {
// 		return err
// 	}
// 	ctx := context.Background()
// 	dt, err := os.ReadFile("./db.sql")
// 	if err != nil {
// 		return err
// 	}

// 	_, err = DB.ExecContext(ctx, string(dt))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
