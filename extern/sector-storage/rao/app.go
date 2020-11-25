package rao

import (
	"fmt"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/api"
	"net/http"
)

func RunWebServer() error {
	http.HandleFunc("/api/v1/updateP2Host",api.UpdateP2HostHandler)
	http.HandleFunc("/api/v1/ping",api.PingHandler)
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		return fmt.Errorf("run rao web server %v \n", err)
	}
	return nil
}
