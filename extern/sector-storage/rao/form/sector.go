package form

type UpdateP2HostFrom struct {
	SectorId string `json:"sectorId"`
	P2Host   string `json:"p2Host"`
}

func (u UpdateP2HostFrom) Valid() (error, bool) {
	if u.SectorId == "" || u.P2Host == "" {
		return nil, false
	}
	return nil, true
}
