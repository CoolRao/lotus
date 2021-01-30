package sectorstorage

func (sh *scheduler) TaskOk(wreq *workerRequest, whnd *workerHandle) bool {
	return TM.TaskOk(wreq, whnd)
}
