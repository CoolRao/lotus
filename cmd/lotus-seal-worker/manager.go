package main

import (
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao"
	"io/ioutil"
)



type Manager struct {
}

func (m *Manager) ReadCfg(path string) (*rao.WorkerConfig, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read worker config file is error,please config %v \n",err)
	}
	workerCfg := rao.WorkerConfig{}
	err = json.Unmarshal(bytes, &workerCfg)
	if err != nil {
		return nil, err
	}
	return &workerCfg, nil
}

func NewManager() *Manager {
	return &Manager{}
}
