package sectorstorage

import (
"encoding/json"
"reflect"
"time"
)

type JobsConfig struct {
	HostName           string
	AddPieceMax        int
	PreCommit1Max      int
	PreCommit2Max      int
	Commit2Max         int
	DiskHoldMax        int
	APDiskHoldMax      int
	ForceP1FromLocalAP bool
	ForceP2FromLocalP1 bool
	ForceC2FromLocalP2 bool
	IsPlanOffline      bool
	AllowP2C2Parallel  bool
	IgnoreOutOfSpace   bool
}

func TimeMinuteSleep(dur int) {
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