package db

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var db *sql.DB
var err error

const dbPath = "/home/abel/lotus.db"

func init() {
	err := SqlInit(dbPath)
	if err != nil {
		panic(err)
	}
}

func SqlInit(dbPath string) error {
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("init db %v \n", err)
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return nil
}

func getSqlDb() *sql.DB {
	return db
}
