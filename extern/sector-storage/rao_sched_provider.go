package sectorstorage

import "github.com/filecoin-project/lotus/extern/sector-storage/sealtasks"

// 这个任务已经分配给确定的worker更新信息
func (sh *scheduler) UpdateSectorInfo(workerId WorkerID, wreq *workerRequest, whnd *workerHandle) {
	jobConfig := whnd.taskInfo.JobConfig
	state := false
	taskType := wreq.taskType
	switch taskType {
	case sealtasks.TTAddPiece:
		state = jobConfig.ForceP1FromLocalAP
		break
	case sealtasks.TTPreCommit1:
		state = jobConfig.ForceP2FromLocalP1
		break
	case sealtasks.TTCommit1:
		state = jobConfig.ForceC2FromLocalP2
		break
	default:

	}
	SM.UpdateSectorInfo(wreq.sector.ID, taskType, whnd.taskInfo.HostName, state)
	whnd.taskInfo.IncrCount(taskType)
	for _, worker := range sh.workers {
		worker.taskInfo.ClearCacheTask(taskType)
	}

}

/*
 true 为倒序，false正序
*/
func (sh *scheduler) TaskCmp(taskType sealtasks.TaskType, a, b *workerHandle) bool {
	aCount := a.taskInfo.GetTotalCount(taskType)
	bCount := b.taskInfo.GetTotalCount(taskType)
	return aCount < bCount

}

func (sh *scheduler) TaskOk(wreq *workerRequest, whnd *workerHandle) bool {

	// 任务数量限制
	ok := whnd.taskInfo.CanAddTask(wreq.taskType)
	log.Infof("rao can add task %v  ",ok)
	if !ok {
		return false
	}
	// 任务关联执行效验
	match := sh.TaskMatch(wreq, whnd)

	if match {
		whnd.taskInfo.IncrCacheCount(wreq.taskType) // 记录任务数，保证任务可以均匀下发
	}
	return match
}

func (sh *scheduler) TaskMatch(wreq *workerRequest, whnd *workerHandle) bool {
	sectorInfo, ok := SM.GetSectorInfo(wreq.sector.ID)
	if !ok {
		return true
	}
	match := sectorInfo.IsMatch(wreq.taskType, whnd.taskInfo.HostName)
	log.Infof("rao task is ok: match: %v, taskType: %v, workerName: %v ,sectorInfo:  %v", match, wreq.taskType, whnd.taskInfo.HostName, sectorInfo.Info())
	return match
}
