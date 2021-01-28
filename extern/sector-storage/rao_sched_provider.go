package sectorstorage

import "github.com/filecoin-project/lotus/extern/sector-storage/sealtasks"

// 这个任务已经分配给确定的worker更新信息
func (sh *scheduler) UpdateSectorInfo(workerId WorkerID, wreq *workerRequest, whnd *workerHandle) {
	jobConfig := whnd.taskInfo.JobsConfig
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
	TM.Update(wreq.sector.ID, taskType, whnd.taskInfo.HostName, state)
	whnd.taskInfo.IncrRunning(taskType)
	for _, worker := range sh.workers {
		worker.taskInfo.ClearCache(taskType)
	}

}

/*
 true 为倒序，false正序
*/
func (sh *scheduler) TaskCmp(taskType sealtasks.TaskType, a, b *workerHandle) bool {
	aCount := a.taskInfo.GetTotal(taskType)
	bCount := b.taskInfo.GetTotal(taskType)
	return aCount < bCount

}

func (sh *scheduler) TaskOk(wreq *workerRequest, whnd *workerHandle) bool {
	return TM.TaskOk(wreq.sector.ID, wreq.taskType, whnd)
}
