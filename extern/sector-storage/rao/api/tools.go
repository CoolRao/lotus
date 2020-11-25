package api

import (
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/log"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/model"
	"io/ioutil"
	"net/http"
	"reflect"
)

func BindJson(req *http.Request, val interface{}) error {
	if reflect.ValueOf(val).Kind() != reflect.Ptr {
		return fmt.Errorf("val is mutst pointer %v \n", val)
	}
	bytes, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, val)
}

func FailResponse(w http.ResponseWriter,msg string)  {
	response := model.Response{Data: nil, Code: 0, Msg: msg}
	bytes, err := json.Marshal(response)
	if err!=nil{
		log.Info("json error %v \n",err)
		return
	}
	_, err = w.Write(bytes)
	if err!=nil{
		log.Info("http write error %v \n",err)
		return
	}
	return
}
func SuccessResponse(w http.ResponseWriter,data interface{})  {
	response := model.Response{Data: data, Code: 0, Msg: "success"}
	bytes, err := json.Marshal(response)
	if err!=nil{
		log.Info("json error %v \n",err)
		return
	}
	_, err = w.Write(bytes)
	if err!=nil{
		log.Info("http write %v \n",err)
		return
	}
	return
}
