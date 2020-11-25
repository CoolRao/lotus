package service

import "github.com/filecoin-project/lotus/extern/sector-storage/rao/model"

type ISectorService interface {
	Update(sectorId, p2Host string) (interface{}, error)
	Get(sectorId string) (*model.Sector, error)
}
