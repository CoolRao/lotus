package sectorstorage

import (
	"fmt"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/extern/sector-storage/sealtasks"

	"sync"
)

type SectorStore map[abi.SectorID]*SectorInfo

func (s SectorStore) Put(sector *SectorInfo) {
	if sector == nil {
		return
	}
	s[sector.ID] = sector
}

func (s SectorStore) Delete(sectorId abi.SectorID) {
	delete(s, sectorId)
}

func (s SectorStore) Get(sectorId abi.SectorID) (*SectorInfo, error) {
	sectorInfo, ok := s[sectorId]
	if ok {
		return sectorInfo, nil
	}
	return nil, fmt.Errorf("sectorStore: no find sector %v ", sectorId)
}

func (s SectorStore) Has(sectorId abi.SectorID) bool {
	_, ok := s[sectorId]
	return ok
}

type CountMap map[sealtasks.TaskType]int

func (c CountMap) Add(taskType sealtasks.TaskType) {
	count, ok := c[taskType]
	if ok {
		c[taskType] = count + 1
	} else {
		c[taskType] = 1
	}
}

func (c CountMap) Get(taskType sealtasks.TaskType) int {
	count, ok := c[taskType]
	if ok {
		return count
	}
	return 0
}

func (c CountMap) Del(taskType sealtasks.TaskType) {
	count, ok := c[taskType]
	if ok {
		c[taskType] = count - 1
	}
}

func (c CountMap) Clear(taskType sealtasks.TaskType) {
	_, ok := c[taskType]
	if ok {
		c[taskType] = 0
	}
}

type TaskCount struct {
	HostName     string
	cacheCount   CountMap
	runningCount CountMap
	JobsConfig   JobsConfig
	lock         sync.RWMutex
}

func NewTaskCount(hostName string, jobCfg map[string]interface{}) *TaskCount {
	jc := JobsConfig{}
	err := MapToStruct(jobCfg, &jc)
	if err != nil {
		panic(err)
	}
	return &TaskCount{HostName: hostName, cacheCount: CountMap{}, runningCount: CountMap{}, JobsConfig: jc}
}

func (w *TaskCount) GetTotal(taskType sealtasks.TaskType) int {
	w.lock.Lock()
	defer w.lock.Unlock()
	cacheCount := w.cacheCount.Get(taskType)
	runningCount := w.runningCount.Get(taskType)
	total := cacheCount + runningCount
	return total
}

func (w *TaskCount) IncrCacheCount(taskType sealtasks.TaskType) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.cacheCount.Add(taskType)
}

func (w *TaskCount) ClearCache(taskType sealtasks.TaskType) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.cacheCount.Clear(taskType)
}

func (w *TaskCount) IncrRunning(taskType sealtasks.TaskType) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.runningCount.Add(taskType)
}

func (w *TaskCount) DelRunning(taskType sealtasks.TaskType) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.runningCount.Del(taskType)
}

var TM = NewTaskManager()

type TaskManager struct {
	sync.RWMutex
	SectorStore SectorStore
}

func NewTaskManager() *TaskManager {
	return &TaskManager{SectorStore: SectorStore{}}
}

func (t TaskManager) TaskOk(sectorId abi.SectorID, taskType sealtasks.TaskType, worker *workerHandle) bool {

	// 判断这个任务是否需要从本地做，如果匹配直接返回 true
	sectorInfo, err := t.Get(sectorId)
	if err != nil {
		log.Error("rao no find sectorInfo: %v \n", err)
	}

	if sectorInfo != nil && !sectorInfo.IsLocalMatch(taskType, worker.info.Hostname) {
		return false
	}

	// 判断这个 worker，在worker对列中,对于该类型的任务数量(已经运行和已经分配的) worker,满足返回  true

	return false
}

func (t TaskManager) Del(sectorId abi.SectorID) {
	t.Lock()
	defer t.Unlock()
	t.SectorStore.Delete(sectorId)
}

func (t TaskManager) Get(sectorId abi.SectorID) (*SectorInfo, error) {
	t.RLock()
	defer t.RUnlock()
	sectorInfo, err := t.SectorStore.Get(sectorId)
	return sectorInfo, err
}

func (t TaskManager) Has(sectorId abi.SectorID) bool {
	t.RLock()
	defer t.RUnlock()
	has := t.SectorStore.Has(sectorId)
	return has
}

func (t TaskManager) Update(sectorId abi.SectorID, taskType sealtasks.TaskType, hostName string, force bool) {
	t.Lock()
	defer t.Unlock()
	has := t.SectorStore.Has(sectorId)
	if !has {
		t.SectorStore.Put(NewSectorInfo(sectorId, taskType))
	} else {
		sectorInfo, err := t.SectorStore.Get(sectorId)
		if err != nil {
			log.Errorf(err.Error())
			return
		}
		sectorInfo.UpdateHost(taskType, hostName, force)
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

func (s *SectorInfo) IsLocalMatch(taskType sealtasks.TaskType, hostName string) bool {
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
	cfg := JobsConfig{}
	err := MapToStruct(jobCfg, &cfg)
	if err != nil {
		log.Errorf("parse job cfg error: %v ", err)
	}
	log.Infof("rao taskManger  wid: %v hostName: %v jobCfg: %v ", wid, hostName, cfg)
	return &WorkerTaskInfo{
		Wid:            wid,
		HostName:       hostName,
		TaskCount:      make(map[sealtasks.TaskType]int),
		CacheTaskCount: make(map[sealtasks.TaskType]int),
		JobConfig:      cfg,
	}
}
