package daoimpl

import (
	"fmt"
	"testing"
)



func test01(sql string,params...interface{}) string{
	return fmt.Sprintf(sql,params...)
}


func TestSql(t *testing.T){
	dao := NewSectorDao()
	info, err := dao.SectorInfo("102")
	fmt.Println(info,err)
}



func TestSectorDao_UpdateInfo(t *testing.T) {
	dao := NewSectorDao()
	info, err := dao.UpdateInfo("1002", "5555")
	fmt.Println(info,err)
}
func TestSectorDao_List(t *testing.T) {
	dao := NewSectorDao()
	list, err := dao.List()
	if err!=nil{
		fmt.Println(err.Error())
		return
	}
	for i:=0;i<len(list);i++{
		sector := list[i]
		fmt.Println(sector.Id,sector.SectorId,sector.P2Host)
	}
}
