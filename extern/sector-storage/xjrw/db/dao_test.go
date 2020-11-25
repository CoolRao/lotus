package db

import (
	"fmt"
	"testing"
)

func TestDaoTotal(t *testing.T) {
	dao := NewDaoWithTable("sectors")
	total, err := dao.Total("select count(*) from sectors")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(total)
}

func TestCreateTable(t *testing.T) {

}

func TestDao_FindOne(t *testing.T) {

}

func TestDao_List(t *testing.T) {

}

func TestDao_Insert(t *testing.T) {

}
