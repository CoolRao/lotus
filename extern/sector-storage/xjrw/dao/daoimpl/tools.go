package daoimpl

import "fmt"

func getSql(sql string, params []interface{}) string {
	return fmt.Sprintf(sql, params)
}

func getParams(params ...interface{}) []interface{} {
	return params
}
