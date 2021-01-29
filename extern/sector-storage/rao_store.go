package sectorstorage

import (
	"context"
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

func (s SectorStore) Get(sectorId abi.SectorID) (*SectorInfo, bool) {
	sectorInfo, ok := s[sectorId]
	if ok {
		return sectorInfo, ok
	}
	return nil, ok
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
	JobsConfig JobsConfig
	TaskCount  CountMap
	WorkerId   WorkerID
	tasks      map[sealtasks.TaskType]struct{}
}

func (w *TaskCount) Support(taskType sealtasks.TaskType) bool {
	_, supported := w.tasks[taskType]
	return supported

}

func NewTaskCount(workerId WorkerID, jobCfg JobsConfig, tasks map[sealtasks.TaskType]struct{}) *TaskCount {
	return &TaskCount{
		TaskCount:  CountMap{},
		WorkerId:   workerId,
		JobsConfig: jobCfg,
		tasks:      tasks,
	}
}

func (w TaskCount) Add(taskType sealtasks.TaskType) {
	w.TaskCount.Add(taskType)
}

func (w TaskCount) Get(taskType sealtasks.TaskType) int {
	return w.TaskCount.Get(taskType)
}

func (w TaskCount) Del(taskType sealtasks.TaskType) {
	w.TaskCount.Del(taskType)
}

func (w TaskCount) TaskCountOk(taskType sealtasks.TaskType) (int, bool) {
	count := w.TaskCount.Get(taskType)
	ok := true
	switch taskType {
	case sealtasks.TTAddPiece:
		if w.JobsConfig.AddPieceMax != 0 {
			ok = count < w.JobsConfig.AddPieceMax
		}
		break
	case sealtasks.TTPreCommit1:
		if w.JobsConfig.PreCommit1Max != 0 {
			ok = count < w.JobsConfig.PreCommit1Max
		}
		break

	case sealtasks.TTPreCommit2:
		if w.JobsConfig.PreCommit2Max != 0 {
			ok = count < w.JobsConfig.PreCommit2Max
		}
		break

	case sealtasks.TTCommit2:
		if w.JobsConfig.Commit2Max != 0 {
			ok = count < w.JobsConfig.Commit2Max
		}
		break

	default:

	}
	return count, ok
}

var TM = NewTaskManager()

type TaskManager struct {
	sync.RWMutex
	SectorStore SectorStore
	TaskCount   map[WorkerID]*TaskCount
}

func NewTaskManager() *TaskManager {
	return &TaskManager{SectorStore: SectorStore{}, TaskCount: make(map[WorkerID]*TaskCount)}
}

func (t TaskManager) TaskOk(task *workerRequest, worker *workerHandle) bool {

	// 判断任务总数是否限制
	workerId := worker.taskInfo.WorkerId
	taskCount, tOK := t.TaskCount[workerId]
	if !tOK {
		types, err := worker.workerRpc.TaskTypes(context.TODO())
		if err != nil {
			log.Errorf("rao taskManager: get taskType error: %v", err)
			return false
		}
		taskCount = NewTaskCount(workerId, worker.taskInfo.JobsConfig, types)
		t.TaskCount[workerId] = taskCount
	}
	count, ok := taskCount.TaskCountOk(task.taskType)
	if !ok {
		log.Infof("rao taskManaer: task is not ok, hostName: %v  count: %v   ", worker.info.Hostname, count)
		return false
	}

	// 判断是否从本地开始做
	sectorId := task.sector.ID
	sectorInfo, sOk := t.SectorStore.Get(sectorId)
	if !sOk {
		sectorInfo = &SectorInfo{ID: sectorId}
		t.SectorStore.Put(sectorInfo)
	}

	if local, match := sectorInfo.NeedFromLocal(task.taskType, worker.info.Hostname); local && !match {
		log.Infof("rao taskManager: sector need from loacal: secororId: %v  taskType: %v  hostName: %v ", sectorId, task.taskType, worker.info.Hostname)
		return false
	}

	// 是否是任务数最小worker 
	minWorker := t.matchMinWorker(task.taskType, count)
	if !minWorker {
		log.Infof("rao taskManager: is not min worker: hostName: %v ", worker.info.Hostname)
		return false
	}

	state := t.GetState(task.taskType, worker.taskInfo.JobsConfig)

	t.Update(sectorId, task.taskType, worker.info.Hostname, state)

	return true
}

func (t *TaskManager) GetState(taskType sealtasks.TaskType, config JobsConfig) bool {
	switch taskType {
	case sealtasks.TTAddPiece:
		return config.ForceP1FromLocalAP

	case sealtasks.TTPreCommit1:
		return config.ForceP2FromLocalP1

	case sealtasks.TTPreCommit2:
		return config.ForceC2FromLocalP2
	default:
	}
	return false
}

func (t *TaskManager) DelSector(workerId WorkerID, taskType sealtasks.TaskType) {
	taskCount, ok := t.TaskCount[workerId]
	if ok {
		taskCount.Del(taskType)
	} else {
		log.Errorf("rao taskManager: hostName: %v taskCount not exists %v ", workerId, taskType)
	}
}

func (t TaskManager) matchMinWorker(taskType sealtasks.TaskType, curCount int) bool {
	min := false
	for _, taskCount := range t.TaskCount {
		if taskCount.Support(taskType) {
			count := taskCount.Get(taskType)
			if curCount <= count {
				min = true
			}
		}
	}
	return min
}

func (t *TaskManager) Update(sectorId abi.SectorID, taskType sealtasks.TaskType, hostName string, force bool) {
	t.Lock()
	defer t.Unlock()
	has := t.SectorStore.Has(sectorId)
	if !has {
		t.SectorStore.Put(NewSectorInfo(sectorId, taskType))
	} else {
		sectorInfo, ok := t.SectorStore.Get(sectorId)
		if !ok {
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

/*
first param:     是否配置了参数
second param:    效验结果
*/
func (s *SectorInfo) NeedFromLocal(taskType sealtasks.TaskType, hostName string) (bool, bool) {
	match := false
	switch taskType {
	case sealtasks.TTPreCommit1:
		if s.ForceP1FromLocalAP {
			match = s.APHost == hostName
			return true, s.APHost == hostName
		}
		break
	case sealtasks.TTPreCommit2:
		if s.ForceP2FromLocalP1 {
			match = s.Pre1Host == hostName
			return true, match
		}
		break
	case sealtasks.TTCommit2:
		if s.ForceC2FromLocalP2 {
			match = s.Pre2Host == hostName
			return true, match
		}
		break
	default:

	}
	return false, match
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
