package dao

import "github.com/filecoin-project/lotus/extern/sector-storage/rao/model"

type ISectorDao interface {
	SectorInfo(sectorId string) (*model.Sector, error)
	Insert(sectorId, p2Host string) (interface{}, error)
	UpdateInfo(sectorId, p2Host string) (interface{}, error)
	List() ([]*model.Sector, error)
}
