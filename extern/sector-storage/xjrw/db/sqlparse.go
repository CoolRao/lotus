package db

import (
	"bytes"
	"fmt"
	"strings"
)

func getQuerySql(sql string, params []interface{}) string {
	if params!=nil && len(params)>0 {
		return fmt.Sprintf(sql, params...)
	}
	return sql
}

func queryTotalSql(sql string) (string, error) {
	srcSql := strings.Split(sql, "limit")
	if len(srcSql) < 1 {
		return "", fmt.Errorf("sql query total")
	}
	resSql := strings.Split(srcSql[0], "from")
	if len(resSql) < 1 {
		return "", fmt.Errorf("ressql no find")
	}
	return fmt.Sprintf("select count(*) from %v", resSql[1]), nil
}

func queryListSql(table string, query ...Param) string {
	if len(query) == 0 {
		return fmt.Sprintf("SELECT * FROM %v ", table)
	}
	var sqlBuff bytes.Buffer
	for key, value := range query[0] {
		tKey := key
		tValue := value
		sqlBuff.WriteString(fmt.Sprintf("%v = %v AND ", tKey, tValue))
	}
	sqlBuff.WriteString(")")
	pageInfo := getPageInfo(query[0])
	if len(query) > 1 {
		//todo
		return ""
	} else {
		return fmt.Sprintf("SELECT * FROM %v WHERE %v %v", table, reviseQueryListSql(sqlBuff.String()), pageInfo)
	}

}

func findOneSql(table string, param Param) string {
	//SELECT column_name,column_name  FROM table_name [WHERE Clause]  [LIMIT N][ OFFSET M]
	var sqlBuff bytes.Buffer
	for key, value := range param {
		tKey := key
		tValue := value
		sqlBuff.WriteString(fmt.Sprintf("%v=%v AND ", tKey, tValue))
	}
	sqlBuff.WriteString(")")
	return fmt.Sprintf("SELECT * FROM %v WHERE %v LIMIT 1", table, reviseQueryListSql(sqlBuff.String()))
}

func insertSql(table string, param Param) (string, []interface{}) {
	// INSERT INTO table_name ( field1, field2,...fieldN ) 	VALUES ( value1, value2,...valueN );
	var keyBuff bytes.Buffer
	var valBuff bytes.Buffer
	keyBuff.WriteString("( ")
	valBuff.WriteString("( ")
	var res []interface{}
	for key, value := range param {
		tKey := key
		tValue := value
		keyBuff.WriteString(fmt.Sprintf("%v, ", tKey))
		valBuff.WriteString("?, ")
		res = append(res, tValue)
	}
	keyBuff.WriteString(")")
	valBuff.WriteString(")")
	return fmt.Sprintf("INSERT INTO %v %v VALUES %v", table, reviseInsertSql(keyBuff.String()), reviseInsertSql(valBuff.String())), res
}

func getPageInfo(query Param) string {
	page, ok := query["page"]
	if !ok {
		return ""
	}
	size, ok := query["size"]
	if ok {
		return ""
	}
	skip := page.(int) * size.(int)
	return fmt.Sprintf("LIMIT %v,%v", skip, size)

}

func reviseInsertSql(src string) string {
	return strings.Replace(src, ", )", " )", 1)
}

func reviseQueryListSql(src string) string {
	return strings.Replace(src, "AND )", "", 1)
}
