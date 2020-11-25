package daoimpl

import (
	"database/sql"
	"fmt"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/dao"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/db"
	"github.com/filecoin-project/lotus/extern/sector-storage/rao/model"
)

type SectorDao struct {
	dao db.IDao
}

func (s *SectorDao) List() ([]*model.Sector, error) {
	var res []*model.Sector
	_, err := s.dao.List("select * from sectors", func(rows *sql.Rows) error {
		for rows.Next() {
			sector := &model.Sector{}
			err := rows.Scan(&sector.Id, &sector.SectorId, &sector.P2Host)
			if err != nil {
				return err
			}
			res = append(res, sector)
		}
		return nil
	})
	return res, err
}

func (s *SectorDao) Insert(sectorId, p2Host string) (interface{}, error) {
	return s.dao.InsertOne("insert into sectors (sectorId,p2Host) values (?, ?)", sectorId, p2Host)
}

func (s *SectorDao) UpdateInfo(sectorId, p2Host string) (interface{}, error) {
	sectorInfo, err := s.SectorInfo(sectorId)
	if err != nil {
		return nil, err
	}
	if sectorInfo.SectorId == "" {
		_, err := s.dao.InsertOne("insert into sectors (sectorId,p2Host) values (?, ?)", sectorId, p2Host)
		if err != nil {
			return nil, fmt.Errorf("insert sector: %v  %v  %v\n", sectorId, p2Host, err)
		}
	} else {
		_, err := s.dao.UpdateOne(`update sectors set p2Host = ? where sectorId = ?`,  p2Host,sectorId)
		if err != nil {
			return nil, fmt.Errorf("update sector: %v %v %v \n", sectorId, p2Host, err)
		}
	}
	return nil, nil
}

func (s *SectorDao) SectorInfo(sectorId string) (*model.Sector, error) {
	sector := &model.Sector{}
	_, err := s.dao.FindOne("select * from sectors where sectorId = %v", func(rows *sql.Rows) error {
		return rows.Scan(&sector.Id, &sector.SectorId, &sector.P1Host)
	}, sectorId)
	if err != nil {
		return nil, err
	}
	return sector, nil
}

func NewSectorDao() dao.ISectorDao {
	return &SectorDao{dao: db.NewDaoWithTable(db.SectorsTable)}
}
