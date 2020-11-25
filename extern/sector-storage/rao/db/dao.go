package db

import (
	"database/sql"
	"fmt"
	"time"
)

type Dao struct {
	db    *sql.DB
	table string
}

func (d *Dao) DeleteOne(sql string, params ...interface{}) (interface{}, error) {
	printSql(sql)
	stmt, err := d.getDb().Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer closeStmt(stmt)
	result, err := stmt.Exec(params)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *Dao) UpdateOne(sql string, params ...interface{}) (interface{}, error) {
	//UPDATE table_name SET field1=new-value1, field2=new-value2 [WHERE Clause]
	printSql(sql)
	stmt, err := d.getDb().Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer closeStmt(stmt)
	result, err := stmt.Exec(params...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *Dao) InsertOne(sql string, params ...interface{}) (interface{}, error) {
	printSql(sql)
	stmt, err := d.getDb().Prepare(sql)
	if err != nil {
		return nil, err
	}
	defer closeStmt(stmt)
	result, err := stmt.Exec(params...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (d *Dao) Total(sql string) (int, error) {
	printSql(sql)
	rows, err := d.getDb().Query(sql)
	if err != nil {
		return 0, err
	}
	defer closeRow(rows)
	total := 0
	for rows.Next() {
		err := rows.Scan(&total)
		if err != nil {
			return 0, err
		}
		return total, nil
	}
	return total, nil
}

func (d *Dao) List(sql string, fn func(row *sql.Rows) error,params ...interface{}) (int, error) {
	querySql := getQuerySql(sql, params)
	printSql(querySql)
	rows, err := d.getDb().Query(querySql)
	if err != nil {
		return 0, err
	}
	defer closeRow(rows)
	err = fn(rows)
	if err != nil {
		return 0, err
	}
	totalSql, err := queryTotalSql(sql)
	if err != nil {
		return 0, fmt.Errorf("get total sql %v \n",err)
	}
	total, err := d.Total(totalSql)
	if err != nil {
		return total, fmt.Errorf("sql total count %v \n",err)
	}
	return total, nil
}

func (d *Dao) UseSession(fn func() error) error {
	tx, err := d.getDb().Begin()
	if err != nil {
		return err
	}
	err = fn()
	if err != nil {
		err := tx.Rollback()
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *Dao) getDb() *sql.DB {
	return getSqlDb()
}


func (d *Dao) FindOne(sql string, fn func(rows *sql.Rows) error, params ...interface{}) (interface{}, error) {
	querySql := getQuerySql(sql, params)
	printSql(querySql)
	rows, err := d.getDb().Query(querySql)
	defer closeRow(rows)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		return nil, fn(rows)
	}
	return nil, nil
}

func closeRow(rows *sql.Rows) {
	if rows != nil {
		err := rows.Close()
		if err != nil {
			fmt.Printf("close rows: %v \n", err)

		}
	}
}

func closeStmt(stmt *sql.Stmt) {
	if stmt != nil {
		err := stmt.Close()
		if err != nil {
			fmt.Printf("close stmt: %v \n", err)
		}
	}
}

func NewDaoWithTable(table string) IDao {
	return &Dao{table: table}
}

func printSql(sql string) {
	fmt.Printf("sqlInfo: %v     %v \n", sql, time.Now())
}
