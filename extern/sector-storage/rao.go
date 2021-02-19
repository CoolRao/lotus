package sectorstorage

import (
	"encoding/json"
	"reflect"
	"time"
)

type JobsConfig struct {
	HostName           string `json:"hostName"`      // 主机名称
	AddPieceMax        int    `json:"addPieceMax"`   // 最大AP个数
	PreCommit1Max      int    `json:"preCommit1Max"` // 最大p1个数
	PreCommit2Max      int    `json:"preCommit2Max"` // 最大p2个数
	Commit2Max         int    `json:"commit2Max"`    // 最大c2个数
	DiskHoldMax        int    `json:"diskHoldMax"`
	APDiskHoldMax      int    `json:"apDiskHoldMax"`
	ForceP1FromLocalAP bool   `json:"forceP1FromLocalAp"` // 强制从本地AP开始p1
	ForceP2FromLocalP1 bool   `json:"forceP2FromLocalP1"` // 强制从本地p1开始p2
	ForceC2FromLocalP2 bool   `json:"forceC2FromLocalP2"` // 强制从本地p2开始c2
	IsPlanOffline      bool   `json:"isPlanOffline"`
	AllowP2C2Parallel  bool   `json:"allowP2C2Parallel"` // 是否允许 p2,c2并行
	IgnoreOutOfSpace   bool   `json:"ignoreOutOfSpace"`  // 是否忽略空间不足提示
}

func getWorkerHostName(defaultHostName string, jobConfig map[string]interface{}) string {
	hostName, ok := jobConfig["HostName"]
	if ok {
		if h, ok1 := hostName.(string); ok1 {
			return h
		}
	}
	return defaultHostName
}



func TimeSleepSecond(dur int) {
	time.Sleep(time.Duration(dur) * time.Second)
}

func MapToStruct(src map[string]interface{}, obj interface{}) error {
	if len(src) == 0 {
		return nil
	}
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return nil
	}
	bytes, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, obj)

}

func StructToMapByReflect(obj interface{}) map[string]interface{} {
	m := make(map[string]interface{})
	if reflect.TypeOf(obj).Kind() != reflect.Ptr {
		return m
	}
	elem := reflect.ValueOf(obj).Elem()
	relType := elem.Type()
	for i := 0; i < relType.NumField(); i++ {
		m[relType.Field(i).Name] = elem.Field(i).Interface()
	}
	return m
}
