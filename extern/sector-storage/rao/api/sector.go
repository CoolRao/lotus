package api

import (
	"fmt"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/form"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/service/impl"
	"log"
	"net/http"
	"time"
)

func UpdateP2HostHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("rao update p2 host: %v  \n", time.Now())
	var updateForm form.UpdateP2HostFrom
	err := BindJson(req, &updateForm)
	if err != nil {
		FailResponse(w, "param not correct")
		return
	}
	sectorService := impl.NewSectorService()
	_, err = sectorService.Update(updateForm.SectorId, updateForm.P2Host)
	if err != nil {
		FailResponse(w, fmt.Sprintf("update sector info %v \n", err))
		return
	}
	SuccessResponse(w, "great")
}

func PingHandler(w http.ResponseWriter, req *http.Request) {
	SuccessResponse(w, fmt.Sprintf("pong is ok %v ", time.Now()))
}
