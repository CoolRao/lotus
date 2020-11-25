package impl

import (
	dao2 "github.com/filecoin-project/lotus/extern/sector-storage/rao/dao"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/dao/daoimpl"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/model"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/service"
)

type SectorService struct {
	dao dao2.ISectorDao
}

func (s *SectorService) Update(sectorId, p2Host string) (interface{}, error) {
	return s.dao.UpdateInfo(sectorId,p2Host)
}

func (s *SectorService) Get(sectorId string) (*model.Sector, error) {
	return s.dao.SectorInfo(sectorId)
}

func NewSectorService() service.ISectorService {
	return &SectorService{dao: daoimpl.NewSectorDao()}
}
