package main

import (
	"encoding/json"
	"fmt"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao"
	"io/ioutil"
)

type Manager struct {
}

func (m *Manager) ReadCfg(path string) (rao.JobsConfig, error) {
	bytes, err := ioutil.ReadFile(path)
	jobsConfig:=rao.JobsConfig{}
	if err != nil {
		return rao.JobsConfig{}, fmt.Errorf("read worker config file is error,please config %v \n", err)
	}
	err = json.Unmarshal(bytes, &jobsConfig)
	if err != nil {
		return rao.JobsConfig{}, err
	}
	return jobsConfig, nil
}

func NewManager() *Manager {
	return &Manager{}
}
