package sectorstorage

import (
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/extern/sector-storage/sealtasks"

	"sync"
)

var SM = NewStorageManager()

type StorageManager struct {
	sectors map[abi.SectorID]*SectorInfo
	sync.Mutex
}

func NewStorageManager() *StorageManager {
	return &StorageManager{
		sectors: make(map[abi.SectorID]*SectorInfo),
	}
}

func (s *StorageManager) GetSectorInfo(sid abi.SectorID) (SectorInfo, bool) {
	s.Lock()
	defer s.Unlock()
	info, ok := s.sectors[sid]
	if ok {
		return *info, true
	}
	return SectorInfo{}, false
}

func (s *StorageManager) UpdateSectorInfo(sid abi.SectorID, taskType sealtasks.TaskType, hostName string, force bool) {
	log.Infof("rao update sector info  sid: %v ,taskType: %v hostName:  %v ", sid, taskType, hostName)
	s.Lock()
	defer s.Unlock()
	sectorInfo, ok := s.sectors[sid]
	if ok {
		sectorInfo.UpdateHost(taskType, hostName, force)
	} else {
		newSectorInfo := NewSectorInfo(sid, taskType)
		newSectorInfo.UpdateHost(taskType, hostName, force)
		s.sectors[sid] = newSectorInfo
	}

}

type SectorInfo struct {
	ID                 abi.SectorID
	TType              sealtasks.TaskType
	APHost             string
	Pre1Host           string
	Pre2Host           string
	C1Host             string
	C2Host             string
	ForceP1FromLocalAP bool
	ForceP2FromLocalP1 bool
	ForceC2FromLocalP2 bool
}

func (s *SectorInfo) Info() string {
	return fmt.Sprintf("Id: %v , ApHost: %v,P1Host: %v ,p2Host: %v ,c2Host: %v ", s.ID, s.APHost, s.Pre1Host, s.Pre2Host, s.C2Host)
}

func (s *SectorInfo) IsMatch(taskType sealtasks.TaskType, hostName string) bool {
	match := true
	switch taskType {
	case sealtasks.TTPreCommit1:
		if s.ForceP1FromLocalAP {
			match = s.APHost == hostName
		}
		break
	case sealtasks.TTPreCommit2:
		if s.ForceP2FromLocalP1 {
			match = s.Pre1Host == hostName
		}
		break
	case sealtasks.TTCommit2:
		if s.ForceC2FromLocalP2 {
			s.Pre2Host = hostName
		}
		break

	default:
	}
	return match
}

func (s *SectorInfo) UpdateHost(taskType sealtasks.TaskType, hostName string, force bool) {
	switch taskType {
	case sealtasks.TTAddPiece:
		s.APHost = hostName
		s.ForceP1FromLocalAP = force
		break
	case sealtasks.TTPreCommit1:
		s.Pre1Host = hostName
		s.ForceP2FromLocalP1 = force
		break
	case sealtasks.TTPreCommit2:
		s.Pre2Host = hostName
		s.ForceC2FromLocalP2 = force
		break
	case sealtasks.TTCommit2:
		s.C2Host = hostName
		break

	default:

	}
}

func NewSectorInfo(Id abi.SectorID, taskType sealtasks.TaskType) *SectorInfo {
	return &SectorInfo{ID: Id, TType: taskType}
}

type WorkerTaskInfo struct {
	Wid            string
	HostName       string
	TaskCount      map[sealtasks.TaskType]int
	CacheTaskCount map[sealtasks.TaskType]int
	JobConfig      JobsConfig
	wl             sync.RWMutex
	cl             sync.RWMutex
	lock           sync.Mutex
}

func (w *WorkerTaskInfo) DelCacheCount(taskType sealtasks.TaskType) {
	w.cl.Lock()
	defer w.cl.Unlock()
	count, ok := w.CacheTaskCount[taskType]
	if ok {
		w.CacheTaskCount[taskType] = count - 1
	}

}

func (w *WorkerTaskInfo) ClearCacheTask(taskType sealtasks.TaskType) {
	w.cl.Lock()
	defer w.cl.Unlock()
	_, ok := w.CacheTaskCount[taskType]
	if ok {
		w.CacheTaskCount[taskType] = 0
	}

}

func (w *WorkerTaskInfo) IncrCacheCount(taskType sealtasks.TaskType) {
	w.cl.Lock()
	defer w.cl.Unlock()
	count, ok := w.CacheTaskCount[taskType]
	if ok {
		w.CacheTaskCount[taskType] = count + 1
	} else {
		w.CacheTaskCount[taskType] = 1
	}

}

func (w *WorkerTaskInfo) GetCacheTaskCount(taskType sealtasks.TaskType) int {
	w.cl.Lock()
	defer w.cl.Unlock()
	count, ok := w.CacheTaskCount[taskType]
	if ok {
		return count
	}
	return 0

}

//--------------运行队列中数量--------------------

func (w *WorkerTaskInfo) IncrCount(taskType sealtasks.TaskType) {
	w.wl.Lock()
	defer w.wl.Unlock()
	count, ok := w.TaskCount[taskType]
	if ok {
		w.TaskCount[taskType] = count + 1
	} else {
		w.TaskCount[taskType] = 1
	}

}

func (w *WorkerTaskInfo) DelCount(taskType sealtasks.TaskType) {
	w.wl.Lock()
	defer w.wl.Unlock()
	count, ok := w.TaskCount[taskType]
	if ok {
		w.TaskCount[taskType] = count - 1
	}

}

func (w *WorkerTaskInfo) GetTaskCount(taskType sealtasks.TaskType) int {
	w.wl.Lock()
	defer w.wl.Unlock()
	count, ok := w.TaskCount[taskType]
	if ok {
		return count
	}
	return 0

}

//-----------------------

func (w *WorkerTaskInfo) GetTotalCount(taskType sealtasks.TaskType) int {
	w.lock.Lock()
	defer w.lock.Unlock()
	count := w.GetTaskCount(taskType) + w.GetCacheTaskCount(taskType)
	return count
}

/**
 * @Description:
 * @receiver w
 * @param taskType
 * @return int  当前该任务的总数
 * @return bool 该任务数量是否进行了跑配置  false 代表没有配置，true代表配置
 */
func (w *WorkerTaskInfo) CmpOk(taskType sealtasks.TaskType) (int, bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	ok := true
	switch taskType {

	case sealtasks.TTAddPiece:
		if w.JobConfig.AddPieceMax == 0 {
			return 0, false
		}

		break
	case sealtasks.TTPreCommit1:
		if w.JobConfig.PreCommit1Max == 0 {
			return 0, false
		}
		break
	case sealtasks.TTPreCommit2:
		if w.JobConfig.PreCommit2Max == 0 {
			return 0, false
		}
		break
	case sealtasks.TTCommit2:
		if w.JobConfig.Commit2Max == 0 {
			return 0, false
		}
		break
	default:
		ok = false
	}
	count := w.GetTaskCount(taskType)

	return count, ok
}

// true 匹配,false不匹配
func (w *WorkerTaskInfo) CanAddTask(taskType sealtasks.TaskType) bool {
	w.lock.Lock()
	defer w.lock.Unlock()
	count := w.GetTotalCount(taskType)
	log.Infof("rao workerName: %v ,cacheCount: %v realCount: %v  taskType: %v", w.HostName, w.GetCacheTaskCount(taskType), w.GetTaskCount(taskType), taskType)
	var ok bool
	switch taskType {
	case sealtasks.TTAddPiece:
		ok = count < w.JobConfig.AddPieceMax
		break
	case sealtasks.TTPreCommit1:
		ok = count < w.JobConfig.PreCommit1Max
		break

	case sealtasks.TTPreCommit2:
		ok = count < w.JobConfig.PreCommit2Max
		break

	case sealtasks.TTCommit2:
		ok = count < w.JobConfig.Commit2Max
		break

	default:
		ok = true
	}
	log.Infof("rao canAddTask workerName: %v  taskType: %v  count: %v  isOK: %v ", w.HostName, taskType, count, ok)
	return ok
}

func newWorkerTaskInfo(wid, hostName string, jobCfg map[string]interface{}) *WorkerTaskInfo {
	log.Infof("rao taskManger  wid: %v hostName: %v jobCfg: %v ", wid, hostName, jobCfg)
	cfg := JobsConfig{}
	err := MapToStruct(jobCfg, &cfg)
	if err != nil {
		log.Errorf("parse job cfg error: %v ", err)
	}
	return &WorkerTaskInfo{
		Wid:            wid,
		HostName:       hostName,
		TaskCount:      make(map[sealtasks.TaskType]int),
		CacheTaskCount: make(map[sealtasks.TaskType]int),
		JobConfig:      cfg,
	}
}
