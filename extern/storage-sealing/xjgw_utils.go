package sealing

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type HttpData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func MarshalJson(request *http.Request, obj interface{}) error {
	bytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return err
	}
	log.Infof("%v request body: %v ", GwLogFilterFlag, string(bytes))
	err = json.Unmarshal(bytes, obj)
	return err
}

func Response(writer http.ResponseWriter, code int, msg string, data interface{}) {
	httpData := HttpData{Code: code, Msg: msg, Data: data}
	bytes, err := json.Marshal(httpData)
	if err != nil {
		log.Errorf("%v http response httpData is error %v ", GwLogFilterFlag, err)
		return
	}
	_, err = writer.Write(bytes)
	if err != nil {
		log.Errorf("%v write body error: %v ", GwLogFilterFlag, err)
	}
}
