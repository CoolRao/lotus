package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/filecoin-project/lotus/extern/sector-storage"
	"io/ioutil"
)




func ReadCfg(path string) (map[string]interface{}, error) {
	bytes, err := ioutil.ReadFile(path)
	configs := make(map[string]interface{})
	if err != nil {
		return nil, fmt.Errorf("read worker config file is error,please config %v \n", err)
	}
	err = json.Unmarshal(bytes, &configs)
	if err != nil {
		return nil, err
	}
	return configs, nil
}

