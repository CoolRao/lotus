package db

import (
	"database/sql"
)

type IDao interface {
	InsertOne(sql string, params ...interface{}) (interface{}, error)
	FindOne(sql string, fn func(rows *sql.Rows) error, params ...interface{}) (interface{}, error)
	List(sql string, fn func(rows *sql.Rows) error,params ...interface{}) (int, error)
	UpdateOne(sql string, params ...interface{}) (interface{}, error)
	DeleteOne(sql string, params ...interface{}) (interface{}, error)
	Total(sql string) (int, error)
	UseSession(func() error) error
}
